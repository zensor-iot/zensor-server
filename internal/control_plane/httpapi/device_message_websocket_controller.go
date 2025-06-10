package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/httpserver"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for development
		// In production, you should validate the origin
		return true
	},
}

type DeviceMessage struct {
	Type      string    `json:"type"`
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

type DeviceMessageWebSocketController struct {
	broker     async.InternalBroker
	clients    map[*websocket.Conn]bool
	clientsMux sync.RWMutex
	broadcast  chan DeviceMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewDeviceMessageWebSocketController(broker async.InternalBroker) *DeviceMessageWebSocketController {
	ctx, cancel := context.WithCancel(context.Background())

	wsc := &DeviceMessageWebSocketController{
		broker:     broker,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan DeviceMessage, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start the hub
	go wsc.run()

	return wsc
}

var _ httpserver.Controller = (*DeviceMessageWebSocketController)(nil)

func (wsc *DeviceMessageWebSocketController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /ws/device-messages", wsc.handleWebSocket())
}

func (wsc *DeviceMessageWebSocketController) handleWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("websocket upgrade failed", slog.String("error", err.Error()))
			return
		}

		slog.Info("new websocket connection established", slog.String("remote_addr", r.RemoteAddr))

		// Register the new client
		wsc.register <- conn

		// Set up ping/pong to keep connection alive
		go wsc.handlePingPong(conn)

		// Handle incoming messages (if any)
		go wsc.handleClient(conn)
	}
}

func (wsc *DeviceMessageWebSocketController) handleClient(conn *websocket.Conn) {
	defer func() {
		wsc.unregister <- conn
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

func (wsc *DeviceMessageWebSocketController) handlePingPong(conn *websocket.Conn) {
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

func (wsc *DeviceMessageWebSocketController) run() {
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

		case client := <-wsc.register:
			wsc.clientsMux.Lock()
			wsc.clients[client] = true
			wsc.clientsMux.Unlock()
			slog.Info("websocket client registered", slog.Int("total_clients", len(wsc.clients)))

		case client := <-wsc.unregister:
			wsc.clientsMux.Lock()
			if _, ok := wsc.clients[client]; ok {
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
			}
			wsc.clientsMux.Unlock()
			slog.Info("websocket client unregistered", slog.Int("total_clients", len(wsc.clients)))

		case message := <-wsc.broadcast:
			wsc.clientsMux.RLock()
			for client := range wsc.clients {
				select {
				case <-wsc.ctx.Done():
					wsc.clientsMux.RUnlock()
					return
				default:
					client.SetWriteDeadline(time.Now().Add(10 * time.Second))
					if err := client.WriteJSON(message); err != nil {
						slog.Error("failed to write message to websocket client", slog.String("error", err.Error()))
						client.Close()
						delete(wsc.clients, client)
					}
				}
			}
			wsc.clientsMux.RUnlock()

		case brokerMsg := <-subscription.Receiver:
			if brokerMsg.Event == "uplink" {
				if envelop, ok := brokerMsg.Value.(dto.Envelop); ok {
					deviceMsg := DeviceMessage{
						Type:      "device_message",
						DeviceID:  envelop.EndDeviceIDs.DeviceID,
						Timestamp: envelop.ReceivedAt,
						Data:      envelop.UplinkMessage.DecodedPayload,
					}

					// Non-blocking send to broadcast channel
					select {
					case wsc.broadcast <- deviceMsg:
					default:
						slog.Warn("broadcast channel full, dropping message")
					}
				}
			}
		}
	}
}

func (wsc *DeviceMessageWebSocketController) Shutdown() {
	slog.Info("shutting down device message websocket controller")
	wsc.cancel()

	wsc.clientsMux.Lock()
	for client := range wsc.clients {
		client.Close()
	}
	wsc.clientsMux.Unlock()

	close(wsc.broadcast)
	close(wsc.register)
	close(wsc.unregister)
}
