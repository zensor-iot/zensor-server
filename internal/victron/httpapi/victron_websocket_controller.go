package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/httpserver"
	victrondto "zensor-server/internal/victron/dto"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type VictronSystemStatusMessage struct {
	Type    string                        `json:"type"`
	Data    victrondto.VictronSystemSnapshot `json:"data"`
	System  VictronSystemSummary           `json:"system"`
}

type VictronSystemSummary struct {
	BatterySOC        float64 `json:"battery_soc"`
	BatteryVoltage    float64 `json:"battery_voltage"`
	BatteryPower      float64 `json:"battery_power"`
	SolarPower        float64 `json:"solar_power"`
	AcLoadPower       float64 `json:"ac_load_power"`
	GridPower         float64 `json:"grid_power"`
	IsCharging        bool    `json:"is_charging"`
	IsInverting       bool    `json:"is_inverting"`
}

type VictronWebSocketController struct {
	broker     async.InternalBroker
	clients    map[*websocket.Conn]bool
	clientsMux sync.RWMutex
	broadcast  chan VictronSystemStatusMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	ctx        context.Context
	cancel     context.CancelFunc
	snapshot   *victrondto.VictronSystemSnapshot
	snapMux    sync.RWMutex
}

func NewVictronWebSocketController(broker async.InternalBroker) *VictronWebSocketController {
	ctx, cancel := context.WithCancel(context.Background())

	wsc := &VictronWebSocketController{
		broker:     broker,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan VictronSystemStatusMessage, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		ctx:        ctx,
		cancel:     cancel,
		snapshot:   &victrondto.VictronSystemSnapshot{},
	}

	go wsc.run()

	return wsc
}

var _ httpserver.Controller = (*VictronWebSocketController)(nil)

func (wsc *VictronWebSocketController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /ws/victron/status", wsc.handleWebSocket())
}

func (wsc *VictronWebSocketController) handleWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("victron websocket upgrade failed", slog.String("error", err.Error()))
			return
		}

		slog.Info("new victron websocket connection established", slog.String("remote_addr", r.RemoteAddr))

		wsc.register <- conn

		go wsc.handlePingPong(conn)
		go wsc.handleClient(conn)
	}
}

func (wsc *VictronWebSocketController) handleClient(conn *websocket.Conn) {
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
				slog.Error("victron websocket read error", slog.String("error", err.Error()))
			} else {
				slog.Debug("victron websocket connection closed", slog.String("error", err.Error()))
			}
			break
		}
	}
}

func (wsc *VictronWebSocketController) handlePingPong(conn *websocket.Conn) {
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

func (wsc *VictronWebSocketController) run() {
	subscription, err := wsc.broker.Subscribe(async.BrokerTopicName("victron_data"))
	if err != nil {
		slog.Error("failed to subscribe to victron data", slog.String("error", err.Error()))
		return
	}
	defer wsc.broker.Unsubscribe(async.BrokerTopicName("victron_data"), subscription)

	for {
		select {
		case <-wsc.ctx.Done():
			return

		case client := <-wsc.register:
			wsc.clientsMux.Lock()
			wsc.clients[client] = true
			wsc.clientsMux.Unlock()
			slog.Info("victron websocket client registered", slog.Int("total_clients", len(wsc.clients)))

			go wsc.sendCurrentSnapshot(client)

		case client := <-wsc.unregister:
			wsc.clientsMux.Lock()
			if _, ok := wsc.clients[client]; ok {
				delete(wsc.clients, client)
				close := func() {
					defer func() {
						if r := recover(); r != nil {
							slog.Warn("recovered from panic while closing victron websocket", slog.Any("panic", r))
						}
					}()
					client.Close()
				}
				close()
			}
			wsc.clientsMux.Unlock()
			slog.Info("victron websocket client unregistered", slog.Int("total_clients", len(wsc.clients)))

		case message := <-wsc.broadcast:
			wsc.clientsMux.RLock()
			clientsToRemove := make([]*websocket.Conn, 0)
			for client := range wsc.clients {
				select {
				case <-wsc.ctx.Done():
					wsc.clientsMux.RUnlock()
					return
				default:
					client.SetWriteDeadline(time.Now().Add(10 * time.Second))
					if err := client.WriteJSON(message); err != nil {
						slog.Error("failed to write victron message to websocket client", slog.String("error", err.Error()))
						clientsToRemove = append(clientsToRemove, client)
					}
				}
			}
			wsc.clientsMux.RUnlock()

			if len(clientsToRemove) > 0 {
				wsc.clientsMux.Lock()
				for _, client := range clientsToRemove {
					if _, ok := wsc.clients[client]; ok {
						delete(wsc.clients, client)
						go func(c *websocket.Conn) {
							defer func() {
								if r := recover(); r != nil {
									slog.Warn("recovered from panic while closing victron websocket", slog.Any("panic", r))
								}
							}()
							c.Close()
						}(client)
					}
				}
				wsc.clientsMux.Unlock()
			}

		case brokerMsg := <-subscription.Receiver:
			if telemetry, ok := brokerMsg.Value.(victrondto.VictronTelemetry); ok {
				wsc.handleTelemetryUpdate(telemetry)
			}
		}
	}
}

func (wsc *VictronWebSocketController) handleTelemetryUpdate(telemetry victrondto.VictronTelemetry) {
	wsc.snapMux.Lock()
	if wsc.snapshot.Raw == nil {
		wsc.snapshot.Raw = make(map[string]victrondto.VictronValue)
	}
	wsc.snapshot.Raw[telemetry.Topic] = telemetry.Value
	wsc.snapshot.PortalID = telemetry.PortalID
	wsc.updateStructuredSnapshot(telemetry)
	snapshotCopy := *wsc.snapshot
	wsc.snapMux.Unlock()

	msg := VictronSystemStatusMessage{
		Type:   "victron_status",
		Data:   snapshotCopy,
		System: buildSummary(snapshotCopy),
	}

	select {
	case wsc.broadcast <- msg:
	default:
		slog.Warn("victron broadcast channel full, dropping message")
	}
}

func (wsc *VictronWebSocketController) updateStructuredSnapshot(telemetry victrondto.VictronTelemetry) {
	switch victrondto.VictronServiceType(telemetry.ServiceType) {
	case victrondto.ServiceBattery:
		wsc.updateBatteryData(telemetry)
	case victrondto.ServiceSolarCharger, victrondto.ServicePvInverter:
		wsc.updateSolarData(telemetry)
	case victrondto.ServiceAcLoad:
		wsc.updateAcLoadData(telemetry)
	case victrondto.ServiceDcLoad:
		wsc.updateDcLoadData(telemetry)
	case victrondto.ServiceVebus:
		wsc.updateVebusData(telemetry)
	case victrondto.ServiceTemperature:
		wsc.updateTemperatureData(telemetry)
	case victrondto.ServiceTank:
		wsc.updateTankData(telemetry)
	}
}

func (wsc *VictronWebSocketController) updateBatteryData(telemetry victrondto.VictronTelemetry) {
	for i, b := range wsc.snapshot.Batteries {
		if b.Instance == telemetry.Instance {
			switch telemetry.Path {
			case "Dc/0/Voltage":
				wsc.snapshot.Batteries[i].Voltage = telemetry.Value.Value
			case "Dc/0/Current":
				wsc.snapshot.Batteries[i].Current = telemetry.Value.Value
			case "Dc/0/Power":
				wsc.snapshot.Batteries[i].Power = telemetry.Value.Value
			case "Soc":
				wsc.snapshot.Batteries[i].Soc = telemetry.Value.Value
			case "Temperature":
				wsc.snapshot.Batteries[i].Temperature = telemetry.Value.Value
			}
			return
		}
	}
	wsc.snapshot.Batteries = append(wsc.snapshot.Batteries, victrondto.BatteryData{
		Instance: telemetry.Instance,
	})
	wsc.updateBatteryData(telemetry)
}

func (wsc *VictronWebSocketController) updateSolarData(telemetry victrondto.VictronTelemetry) {
	for i, s := range wsc.snapshot.SolarChargers {
		if s.Instance == telemetry.Instance {
			switch telemetry.Path {
			case "Dc/0/Power":
				wsc.snapshot.SolarChargers[i].Power = telemetry.Value.Value
			case "Dc/0/Voltage":
				wsc.snapshot.SolarChargers[i].Voltage = telemetry.Value.Value
			case "Dc/0/Current":
				wsc.snapshot.SolarChargers[i].Current = telemetry.Value.Value
			case "Yield/Power":
				wsc.snapshot.SolarChargers[i].Power = telemetry.Value.Value
			}
			return
		}
	}
	wsc.snapshot.SolarChargers = append(wsc.snapshot.SolarChargers, victrondto.SolarChargerData{
		Instance: telemetry.Instance,
	})
	wsc.updateSolarData(telemetry)
}

func (wsc *VictronWebSocketController) updateAcLoadData(telemetry victrondto.VictronTelemetry) {
	for i, l := range wsc.snapshot.AcLoads {
		if l.Instance == telemetry.Instance {
			switch telemetry.Path {
			case "Ac/0/Power":
				wsc.snapshot.AcLoads[i].Power = telemetry.Value.Value
			case "Ac/0/Voltage":
				wsc.snapshot.AcLoads[i].Voltage = telemetry.Value.Value
			case "Ac/0/Current":
				wsc.snapshot.AcLoads[i].Current = telemetry.Value.Value
			}
			return
		}
	}
	wsc.snapshot.AcLoads = append(wsc.snapshot.AcLoads, victrondto.LoadData{
		Instance: telemetry.Instance,
	})
	wsc.updateAcLoadData(telemetry)
}

func (wsc *VictronWebSocketController) updateDcLoadData(telemetry victrondto.VictronTelemetry) {
	for i, l := range wsc.snapshot.DcLoads {
		if l.Instance == telemetry.Instance {
			switch telemetry.Path {
			case "Dc/0/Power":
				wsc.snapshot.DcLoads[i].Power = telemetry.Value.Value
			case "Dc/0/Voltage":
				wsc.snapshot.DcLoads[i].Voltage = telemetry.Value.Value
			case "Dc/0/Current":
				wsc.snapshot.DcLoads[i].Current = telemetry.Value.Value
			}
			return
		}
	}
	wsc.snapshot.DcLoads = append(wsc.snapshot.DcLoads, victrondto.LoadData{
		Instance: telemetry.Instance,
	})
	wsc.updateDcLoadData(telemetry)
}

func (wsc *VictronWebSocketController) updateVebusData(telemetry victrondto.VictronTelemetry) {
	for i, v := range wsc.snapshot.Vebus {
		if v.Instance == telemetry.Instance {
			switch telemetry.Path {
			case "Ac/State":
				wsc.snapshot.Vebus[i].State = telemetry.Value.Value
			case "Ac/Out/P":
				wsc.snapshot.Vebus[i].Power = telemetry.Value.Value
			}
			return
		}
	}
	wsc.snapshot.Vebus = append(wsc.snapshot.Vebus, victrondto.VebusData{
		Instance: telemetry.Instance,
	})
	wsc.updateVebusData(telemetry)
}

func (wsc *VictronWebSocketController) updateTemperatureData(telemetry victrondto.VictronTelemetry) {
	for i, t := range wsc.snapshot.Temperatures {
		if t.Instance == telemetry.Instance {
			wsc.snapshot.Temperatures[i].Temperature = telemetry.Value.Value
			return
		}
	}
	wsc.snapshot.Temperatures = append(wsc.snapshot.Temperatures, victrondto.TemperatureData{
		Instance:    telemetry.Instance,
		Temperature: telemetry.Value.Value,
	})
}

func (wsc *VictronWebSocketController) updateTankData(telemetry victrondto.VictronTelemetry) {
	for i, t := range wsc.snapshot.Tanks {
		if t.Instance == telemetry.Instance {
			wsc.snapshot.Tanks[i].Level = telemetry.Value.Value
			return
		}
	}
	wsc.snapshot.Tanks = append(wsc.snapshot.Tanks, victrondto.TankData{
		Instance: telemetry.Instance,
		Level:    telemetry.Value.Value,
	})
}

func (wsc *VictronWebSocketController) sendCurrentSnapshot(client *websocket.Conn) {
	wsc.snapMux.RLock()
	snapshotCopy := *wsc.snapshot
	wsc.snapMux.RUnlock()

	msg := VictronSystemStatusMessage{
		Type:   "victron_status",
		Data:   snapshotCopy,
		System: buildSummary(snapshotCopy),
	}

	client.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := client.WriteJSON(msg); err != nil {
		slog.Error("failed to send victron snapshot to new client", slog.String("error", err.Error()))
	}
}

func buildSummary(snapshot victrondto.VictronSystemSnapshot) VictronSystemSummary {
	summary := VictronSystemSummary{}

	for _, b := range snapshot.Batteries {
		summary.BatterySOC = b.Soc
		summary.BatteryVoltage = b.Voltage
		summary.BatteryPower += b.Power
		if b.Power > 0 {
			summary.IsCharging = true
		}
	}

	for _, s := range snapshot.SolarChargers {
		summary.SolarPower += s.Power
	}

	for _, p := range snapshot.PvInverters {
		summary.SolarPower += p.Power
	}

	for _, l := range snapshot.AcLoads {
		summary.AcLoadPower += l.Power
	}

	gridPower := summary.AcLoadPower - summary.SolarPower - summary.BatteryPower
	if gridPower > 0 {
		summary.GridPower = gridPower
	}
	if summary.BatteryPower < 0 {
		summary.IsInverting = true
	}

	return summary
}

func (wsc *VictronWebSocketController) Shutdown() {
	slog.Info("shutting down victron websocket controller")
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
