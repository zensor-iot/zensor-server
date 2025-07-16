package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/gorilla/websocket"
)

// DeviceSpecificMessage represents a message sent to a device-specific WebSocket client
type DeviceSpecificMessage struct {
	Type      string    `json:"type"`
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

// DeviceSpecificStateMessage represents a device state message for device-specific WebSocket
type DeviceSpecificStateMessage struct {
	Type      string                           `json:"type"`
	DeviceID  string                           `json:"device_id"`
	Timestamp time.Time                        `json:"timestamp"`
	Data      map[string][]usecases.SensorData `json:"data"`
}

// ClientSubscription represents a WebSocket client subscription to a specific device
type ClientSubscription struct {
	conn     *websocket.Conn
	deviceID string
}

// DeviceSpecificWebSocketController handles WebSocket connections for device-specific notifications
type DeviceSpecificWebSocketController struct {
	broker     async.InternalBroker
	stateCache usecases.DeviceStateCacheService
	clients    map[*websocket.Conn]*ClientSubscription
	clientsMux sync.RWMutex
	register   chan *ClientSubscription
	unregister chan *websocket.Conn
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewDeviceSpecificWebSocketController creates a new device-specific WebSocket controller
func NewDeviceSpecificWebSocketController(broker async.InternalBroker, stateCache usecases.DeviceStateCacheService) *DeviceSpecificWebSocketController {
	ctx, cancel := context.WithCancel(context.Background())

	wsc := &DeviceSpecificWebSocketController{
		broker:     broker,
		stateCache: stateCache,
		clients:    make(map[*websocket.Conn]*ClientSubscription),
		register:   make(chan *ClientSubscription),
		unregister: make(chan *websocket.Conn),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start the hub
	go wsc.run()

	return wsc
}

var _ httpserver.Controller = (*DeviceSpecificWebSocketController)(nil)

func (wsc *DeviceSpecificWebSocketController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /ws/devices/{device_id}/messages", wsc.handleWebSocket())
}

func (wsc *DeviceSpecificWebSocketController) handleWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract device ID from URL path
		deviceID := r.PathValue("device_id")
		if deviceID == "" || strings.TrimSpace(deviceID) == "" {
			http.Error(w, "device_id is required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("websocket upgrade failed", slog.String("error", err.Error()))
			return
		}

		slog.Info("new device-specific websocket connection established",
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("device_id", deviceID))

		// Create client subscription
		subscription := &ClientSubscription{
			conn:     conn,
			deviceID: deviceID,
		}

		// Register the new client
		wsc.register <- subscription

		// Set up ping/pong to keep connection alive
		go wsc.handlePingPong(conn)

		// Handle incoming messages (if any)
		go wsc.handleClient(conn)
	}
}

func (wsc *DeviceSpecificWebSocketController) handleClient(conn *websocket.Conn) {
	defer func() {
		// Check if context is done before trying to send to channel
		select {
		case <-wsc.ctx.Done():
			// Controller is shutting down, just close the connection
		default:
			// Try to unregister the client
			select {
			case wsc.unregister <- conn:
			default:
				// Channel might be full, but that's okay
			}
		}
		conn.Close()
	}()

	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error", slog.String("error", err.Error()))
			} else {
				// Log normal closures as debug instead of error
				slog.Debug("websocket connection closed", slog.String("error", err.Error()))
			}
			break
		}
	}
}

func (wsc *DeviceSpecificWebSocketController) handlePingPong(conn *websocket.Conn) {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wsc.ctx.Done():
			return
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (wsc *DeviceSpecificWebSocketController) run() {
	// Subscribe to device messages
	subscription, err := wsc.broker.Subscribe(async.BrokerTopicName("device_messages"))
	if err != nil {
		slog.Error("failed to subscribe to device messages", slog.String("error", err.Error()))
		return
	}
	defer wsc.broker.Unsubscribe(async.BrokerTopicName("device_messages"), subscription)

	for {
		select {
		case <-wsc.ctx.Done():
			return

		case clientSub := <-wsc.register:
			wsc.clientsMux.Lock()
			wsc.clients[clientSub.conn] = clientSub
			wsc.clientsMux.Unlock()
			slog.Info("device-specific websocket client registered",
				slog.String("device_id", clientSub.deviceID),
				slog.Int("total_clients", len(wsc.clients)))

			// Send cached device state to the new client if available
			go wsc.sendCachedStateToClient(clientSub)

		case client := <-wsc.unregister:
			wsc.clientsMux.Lock()
			if clientSub, ok := wsc.clients[client]; ok {
				delete(wsc.clients, client)
				close := func() {
					defer func() {
						if r := recover(); r != nil {
							slog.Warn("recovered from panic while closing websocket", slog.Any("panic", r))
						}
					}()
					client.Close()
				}
				close()
				slog.Info("device-specific websocket client unregistered",
					slog.String("device_id", clientSub.deviceID),
					slog.Int("total_clients", len(wsc.clients)))
			}
			wsc.clientsMux.Unlock()

		case brokerMsg := <-subscription.Receiver:
			if brokerMsg.Event == "uplink" {
				if envelop, ok := brokerMsg.Value.(dto.Envelop); ok {
					deviceMsg := DeviceSpecificMessage{
						Type:      "device_state",
						DeviceID:  envelop.EndDeviceIDs.DeviceID,
						Timestamp: envelop.ReceivedAt,
						Data:      envelop.UplinkMessage.DecodedPayload,
					}

					// Send message only to clients subscribed to this specific device
					wsc.sendMessageToDeviceClients(envelop.EndDeviceIDs.DeviceID, deviceMsg)
				}
			} else if brokerMsg.Event == "command_sent" {
				// Handle command sent events if needed
				if command, ok := brokerMsg.Value.(domain.Command); ok {
					deviceMsg := DeviceSpecificMessage{
						Type:      "command_sent",
						DeviceID:  command.Device.ID.String(),
						Timestamp: command.CreatedAt.Time,
						Data:      command.Payload,
					}

					wsc.sendMessageToDeviceClients(command.Device.ID.String(), deviceMsg)
				}
			}
		}
	}
}

func (wsc *DeviceSpecificWebSocketController) sendMessageToDeviceClients(deviceID string, message DeviceSpecificMessage) {
	wsc.clientsMux.RLock()
	clientsToRemove := make([]*websocket.Conn, 0)
	for conn, clientSub := range wsc.clients {
		if clientSub.deviceID == deviceID {
			select {
			case <-wsc.ctx.Done():
				wsc.clientsMux.RUnlock()
				return
			default:
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteJSON(message); err != nil {
					slog.Error("failed to write message to device-specific websocket client",
						slog.String("device_id", deviceID),
						slog.String("error", err.Error()))
					clientsToRemove = append(clientsToRemove, conn)
				}
			}
		}
	}
	wsc.clientsMux.RUnlock()

	// Remove failed clients with proper synchronization
	if len(clientsToRemove) > 0 {
		wsc.clientsMux.Lock()
		for _, conn := range clientsToRemove {
			if _, ok := wsc.clients[conn]; ok {
				delete(wsc.clients, conn)
				// Close connection in a safe way
				go func(c *websocket.Conn) {
					defer func() {
						if r := recover(); r != nil {
							slog.Warn("recovered from panic while closing websocket", slog.Any("panic", r))
						}
					}()
					c.Close()
				}(conn)
			}
		}
		wsc.clientsMux.Unlock()
	}
}

func (wsc *DeviceSpecificWebSocketController) sendCachedStateToClient(clientSub *ClientSubscription) {
	slog.Info("sending cached state to new device-specific client",
		slog.String("remote_addr", clientSub.conn.RemoteAddr().String()),
		slog.String("device_id", clientSub.deviceID))

	// Get cached state for the specific device
	state, exists := wsc.stateCache.GetState(context.Background(), clientSub.deviceID)
	if !exists {
		slog.Debug("no cached state found for device", slog.String("device_id", clientSub.deviceID))
		return
	}

	slog.Info("sending cached state for device",
		slog.String("device_id", clientSub.deviceID),
		slog.Int("sensor_types", len(state.Data)))

	stateMsg := DeviceSpecificStateMessage{
		Type:      "device_state",
		DeviceID:  clientSub.deviceID,
		Timestamp: state.Timestamp,
		Data:      state.Data,
	}

	clientSub.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := clientSub.conn.WriteJSON(stateMsg); err != nil {
		slog.Error("failed to send cached state to new device-specific client",
			slog.String("device_id", clientSub.deviceID),
			slog.String("error", err.Error()))
		return
	}

	slog.Info("finished sending cached state to new device-specific client",
		slog.String("remote_addr", clientSub.conn.RemoteAddr().String()),
		slog.String("device_id", clientSub.deviceID))
}

func (wsc *DeviceSpecificWebSocketController) Shutdown() {
	slog.Info("shutting down device-specific websocket controller")
	wsc.cancel()

	wsc.clientsMux.Lock()
	for client := range wsc.clients {
		client.Close()
	}
	wsc.clientsMux.Unlock()

	close(wsc.register)
	close(wsc.unregister)
}
