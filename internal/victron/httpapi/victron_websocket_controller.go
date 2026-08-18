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

const (
	victronStatusMessageType = "victron_status"
	noVictronDataErrMessage  = "no victron data received yet"

	// acInputConnectedVoltage is the threshold below which the AC input is
	// treated as disconnected, used only when the GX does not publish
	// Ac/ActiveIn/Connected. A live input sits near nominal line voltage;
	// a disconnected one collapses well below this.
	acInputConnectedVoltage = 100

	// acActiveInputNone is the value Victron uses for Ac/ActiveIn/ActiveInput
	// and Ac/ActiveIn/Source to mean that no AC input is connected.
	acActiveInputNone = 240

	victronWebSocketPingInterval = 54 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type VictronSystemStatusMessage struct {
	Type   string                           `json:"type"`
	Data   victrondto.VictronSystemSnapshot `json:"data"`
	System VictronSystemSummary             `json:"system"`
}

type VictronSystemSummary struct {
	// BatterySOC is nil until a state of charge is reported, so that an
	// unknown charge is not indistinguishable from a flat battery.
	BatterySOC     *float64 `json:"battery_soc"`
	BatteryVoltage float64  `json:"battery_voltage"`
	BatteryPower   float64  `json:"battery_power"`
	SolarPower     float64  `json:"solar_power"`
	AcLoadPower    float64  `json:"ac_load_power"`
	GridPower      float64  `json:"grid_power"`
	IsCharging     bool     `json:"is_charging"`
	IsInverting    bool     `json:"is_inverting"`

	AcInputVoltage   float64 `json:"ac_input_voltage"`
	AcInputConnected bool    `json:"ac_input_connected"`
}

type VictronWebSocketController struct {
	broker       async.InternalBroker
	clients      map[*websocket.Conn]bool
	clientsMux   sync.RWMutex
	broadcast    chan VictronSystemStatusMessage
	register     chan *websocket.Conn
	unregister   chan *websocket.Conn
	ctx          context.Context
	cancel       context.CancelFunc
	snapshot     *victrondto.VictronSystemSnapshot
	hasData      bool
	snapMux      sync.RWMutex
	pingInterval time.Duration
}

func NewVictronWebSocketController(broker async.InternalBroker) *VictronWebSocketController {
	return newVictronWebSocketController(broker, victronWebSocketPingInterval)
}

func NewVictronWebSocketControllerWithPingInterval(broker async.InternalBroker, pingInterval time.Duration) *VictronWebSocketController {
	return newVictronWebSocketController(broker, pingInterval)
}

func newVictronWebSocketController(broker async.InternalBroker, pingInterval time.Duration) *VictronWebSocketController {
	ctx, cancel := context.WithCancel(context.Background())

	wsc := &VictronWebSocketController{
		broker:       broker,
		clients:      make(map[*websocket.Conn]bool),
		broadcast:    make(chan VictronSystemStatusMessage, 256),
		register:     make(chan *websocket.Conn),
		unregister:   make(chan *websocket.Conn),
		ctx:          ctx,
		cancel:       cancel,
		snapshot:     &victrondto.VictronSystemSnapshot{},
		pingInterval: pingInterval,
	}

	go wsc.run()

	return wsc
}

var _ httpserver.Controller = (*VictronWebSocketController)(nil)

func (wsc *VictronWebSocketController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /ws/victron/status", wsc.handleWebSocket())
	router.Handle("GET /v1/victron/latest", wsc.getLatest())
	router.Handle("GET /v1/victron/summary", wsc.getSummary())
}

func (wsc *VictronWebSocketController) getLatest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot, ok := wsc.currentSnapshot()
		if !ok {
			httpserver.ReplyWithError(w, http.StatusServiceUnavailable, noVictronDataErrMessage)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, VictronSystemStatusMessage{
			Type:   victronStatusMessageType,
			Data:   snapshot,
			System: buildSummary(snapshot),
		})
	}
}

func (wsc *VictronWebSocketController) getSummary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot, ok := wsc.currentSnapshot()
		if !ok {
			httpserver.ReplyWithError(w, http.StatusServiceUnavailable, noVictronDataErrMessage)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, buildSummary(snapshot))
	}
}

func (wsc *VictronWebSocketController) currentSnapshot() (victrondto.VictronSystemSnapshot, bool) {
	wsc.snapMux.RLock()
	defer wsc.snapMux.RUnlock()

	return *wsc.snapshot, wsc.hasData
}

func (wsc *VictronWebSocketController) handleWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("victron websocket upgrade failed", slog.String("error", err.Error()))
			return
		}

		slog.Info("new victron websocket connection established", slog.String("remote_addr", r.RemoteAddr))

		select {
		case wsc.register <- conn:
		case <-wsc.ctx.Done():
			conn.Close()
			return
		}

		go wsc.handleClient(conn)
	}
}

func (wsc *VictronWebSocketController) handleClient(conn *websocket.Conn) {
	defer func() {
		select {
		case wsc.unregister <- conn:
		case <-wsc.ctx.Done():
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
				slog.Error("victron websocket read error", slog.String("error", err.Error()))
			} else {
				slog.Debug("victron websocket connection closed", slog.String("error", err.Error()))
			}
			break
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

	pingTicker := time.NewTicker(wsc.pingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-wsc.ctx.Done():
			return

		case client := <-wsc.register:
			wsc.clientsMux.Lock()
			wsc.clients[client] = true
			wsc.clientsMux.Unlock()
			slog.Info("victron websocket client registered", slog.Int("total_clients", len(wsc.clients)))

			// Sent synchronously: this is the only goroutine that writes to
			// client conns (gorilla/websocket forbids concurrent writers),
			// so this must not run concurrently with the broadcast case below.
			wsc.sendCurrentSnapshot(client)

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
			wsc.writeToAllClients(func(client *websocket.Conn) error {
				return client.WriteJSON(message)
			})

		case <-pingTicker.C:
			wsc.writeToAllClients(func(client *websocket.Conn) error {
				return client.WriteMessage(websocket.PingMessage, nil)
			})

		case brokerMsg := <-subscription.Receiver:
			if telemetry, ok := brokerMsg.Value.(victrondto.VictronTelemetry); ok {
				wsc.handleTelemetryUpdate(telemetry)
			}
		}
	}
}

func (wsc *VictronWebSocketController) writeToAllClients(write func(*websocket.Conn) error) {
	wsc.clientsMux.RLock()
	clientsToRemove := make([]*websocket.Conn, 0)
	for client := range wsc.clients {
		select {
		case <-wsc.ctx.Done():
			wsc.clientsMux.RUnlock()
			return
		default:
			client.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := write(client); err != nil {
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
}

func (wsc *VictronWebSocketController) handleTelemetryUpdate(telemetry victrondto.VictronTelemetry) {
	wsc.snapMux.Lock()
	if wsc.snapshot.Raw == nil {
		wsc.snapshot.Raw = make(map[string]victrondto.VictronValue)
	}
	wsc.snapshot.Raw[telemetry.Topic] = telemetry.Value
	wsc.snapshot.PortalID = telemetry.PortalID
	wsc.updateStructuredSnapshot(telemetry)
	wsc.hasData = true
	snapshotCopy := *wsc.snapshot
	wsc.snapMux.Unlock()

	msg := VictronSystemStatusMessage{
		Type:   victronStatusMessageType,
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
	case victrondto.ServiceSystem:
		wsc.updateSystemData(telemetry)
	}
}

func (wsc *VictronWebSocketController) updateSystemData(telemetry victrondto.VictronTelemetry) {
	switch telemetry.Path {
	case "Ac/Grid/L1/Power":
		wsc.snapshot.System.GridL1Power = telemetry.Value.Value
	case "Ac/Grid/L2/Power":
		wsc.snapshot.System.GridL2Power = telemetry.Value.Value
	case "Ac/Grid/L3/Power":
		wsc.snapshot.System.GridL3Power = telemetry.Value.Value
	case "Ac/Consumption/L1/Power":
		wsc.snapshot.System.ConsumptionL1Power = telemetry.Value.Value
	case "Ac/Consumption/L2/Power":
		wsc.snapshot.System.ConsumptionL2Power = telemetry.Value.Value
	case "Ac/Consumption/L3/Power":
		wsc.snapshot.System.ConsumptionL3Power = telemetry.Value.Value
	case "Dc/Battery/Soc":
		wsc.snapshot.System.BatterySoc = telemetry.Value.Value
	case "Dc/Battery/Power":
		wsc.snapshot.System.BatteryPower = telemetry.Value.Value
	case "Dc/Pv/Power":
		wsc.snapshot.System.PvPower = telemetry.Value.Value
	case "Ac/ActiveIn/Source":
		source := telemetry.Value.Value
		wsc.snapshot.System.AcActiveInSource = &source
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
			case "Ac/ActiveIn/L1/V":
				wsc.snapshot.Vebus[i].ActiveInVoltage = telemetry.Value.Value
			case "Ac/ActiveIn/L1/I":
				wsc.snapshot.Vebus[i].ActiveInCurrent = telemetry.Value.Value
			case "Ac/ActiveIn/Connected":
				connected := telemetry.Value.Value > 0
				wsc.snapshot.Vebus[i].ActiveInConnected = &connected
			case "Ac/ActiveIn/ActiveInput":
				input := telemetry.Value.Value
				wsc.snapshot.Vebus[i].ActiveInput = &input
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
		Type:   victronStatusMessageType,
		Data:   snapshotCopy,
		System: buildSummary(snapshotCopy),
	}

	client.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := client.WriteJSON(msg); err != nil {
		slog.Error("failed to send victron snapshot to new client", slog.String("error", err.Error()))
	}
}

// acInputState reports the AC input voltage and whether that input is actually
// feeding the system. The measured voltage is not evidence on its own: the
// inverter keeps measuring nominal line voltage while its transfer relay is
// open, so any explicit indicator from the GX wins over the threshold.
func acInputState(snapshot victrondto.VictronSystemSnapshot) (float64, bool) {
	var voltage float64
	var connected *bool
	var activeInput *float64

	for _, v := range snapshot.Vebus {
		if v.ActiveInVoltage != 0 {
			voltage = v.ActiveInVoltage
		}
		if v.ActiveInConnected != nil {
			connected = v.ActiveInConnected
		}
		if v.ActiveInput != nil {
			activeInput = v.ActiveInput
		}
	}

	switch {
	case connected != nil:
		return voltage, *connected
	case activeInput != nil:
		return voltage, *activeInput != acActiveInputNone
	case snapshot.System.AcActiveInSource != nil:
		return voltage, *snapshot.System.AcActiveInSource != acActiveInputNone
	default:
		return voltage, voltage >= acInputConnectedVoltage
	}
}

func buildSummary(snapshot victrondto.VictronSystemSnapshot) VictronSystemSummary {
	summary := VictronSystemSummary{}

	for _, b := range snapshot.Batteries {
		if b.Soc != 0 {
			soc := b.Soc
			summary.BatterySOC = &soc
		}
		if b.Voltage != 0 {
			summary.BatteryVoltage = b.Voltage
		}
		summary.BatteryPower += b.Power
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

	summary.AcInputVoltage, summary.AcInputConnected = acInputState(snapshot)

	// The GX device's own "system" service publishes pre-aggregated totals
	// across all connected devices; prefer these over the per-device sums
	// above (which only cover devices that expose their own MQTT service).
	sys := snapshot.System
	if consumption := sys.ConsumptionL1Power + sys.ConsumptionL2Power + sys.ConsumptionL3Power; consumption != 0 {
		summary.AcLoadPower = consumption
	}
	if sys.BatterySoc != 0 {
		soc := sys.BatterySoc
		summary.BatterySOC = &soc
	}
	if sys.BatteryPower != 0 {
		summary.BatteryPower = sys.BatteryPower
	}
	if sys.PvPower != 0 {
		summary.SolarPower = sys.PvPower
	}

	// Without a grid meter the GX reports no grid figures at all, so the
	// balance of load, solar and battery is the only estimate available. It
	// is only meaningful while the AC input is actually feeding the system;
	// otherwise it would report the battery discharge as grid import.
	if grid := sys.GridL1Power + sys.GridL2Power + sys.GridL3Power; grid != 0 {
		summary.GridPower = grid
	} else if summary.AcInputConnected {
		if gridPower := summary.AcLoadPower - summary.SolarPower - summary.BatteryPower; gridPower > 0 {
			summary.GridPower = gridPower
		}
	}

	if summary.BatteryPower > 0 {
		summary.IsCharging = true
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
}
