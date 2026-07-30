package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/mqtt"
	victrondto "zensor-server/internal/victron/dto"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	BrokerTopicVictronData async.BrokerTopicName = "victron_data"
	keepAliveInterval                            = 30 * time.Second
)

func NewVictronWorker(
	portalID string,
	mqttClient mqtt.Client,
	broker async.InternalBroker,
) *VictronWorker {
	return &VictronWorker{
		portalID:   portalID,
		mqttClient: mqttClient,
		broker:     broker,
		snapshot:   &VictronSystemState{},
	}
}

var _ async.Worker = (*VictronWorker)(nil)

type VictronSystemState struct {
	mu       sync.RWMutex
	snapshot victrondto.VictronSystemSnapshot
}

func (s *VictronSystemState) Update(telemetry victrondto.VictronTelemetry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.snapshot.Raw == nil {
		s.snapshot.Raw = make(map[string]victrondto.VictronValue)
	}
	s.snapshot.Raw[telemetry.Topic] = telemetry.Value
	s.snapshot.PortalID = telemetry.PortalID
}

func (s *VictronSystemState) GetSnapshot() victrondto.VictronSystemSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

type VictronWorker struct {
	portalID   string
	mqttClient mqtt.Client
	broker     async.InternalBroker
	snapshot   *VictronSystemState
}

func (w *VictronWorker) Run(ctx context.Context, done func()) {
	slog.Debug("victron worker starting",
		slog.String("portal_id", w.portalID),
	)
	defer done()

	topic := fmt.Sprintf("N/%s/#", w.portalID)
	slog.Info("subscribing to victron telemetry",
		slog.String("topic", topic),
	)

	err := w.mqttClient.Subscribe(topic, 0, w.handleMessage)
	if err != nil {
		slog.Error("subscribing to victron telemetry",
			slog.String("topic", topic),
			slog.Any("error", err),
		)
		return
	}

	// Venus OS only keeps streaming values on N/<portalID>/# while it
	// receives a periodic keep-alive; otherwise it goes quiet after
	// the initial retained burst. See victronenergy/dbus-mqtt.
	w.sendKeepAlive(ctx)

	ticker := time.NewTicker(keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("victron worker cancelled")
			return
		case <-ticker.C:
			w.sendKeepAlive(ctx)
		}
	}
}

func (w *VictronWorker) sendKeepAlive(ctx context.Context) {
	topic := fmt.Sprintf("R/%s/keepalive", w.portalID)
	if err := w.mqttClient.Publish(ctx, topic, ""); err != nil {
		slog.Error("publishing victron keepalive",
			slog.String("topic", topic),
			slog.Any("error", err),
		)
	}
}

func (w *VictronWorker) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	telemetry, err := victrondto.ParseTelemetry(msg.Topic(), msg.Payload())
	if err != nil {
		slog.Error("parsing victron telemetry",
			slog.String("topic", msg.Topic()),
			slog.Any("error", err),
		)
		return
	}

	slog.Debug("victron telemetry received",
		slog.String("service_type", telemetry.ServiceType),
		slog.Int("instance", telemetry.Instance),
		slog.String("path", telemetry.Path),
		slog.Float64("value", telemetry.Value.Value),
	)

	w.snapshot.Update(telemetry)

	eventName := fmt.Sprintf("%s_%s",
		victrondto.ToSnakeName(telemetry.ServiceType),
		normalizePath(telemetry.Path),
	)

	brokerMsg := async.BrokerMessage{
		Event: eventName,
		Value: telemetry,
	}

	ctx := context.Background()
	tracer := otel.Tracer("zensor_server")
	ctx, span := tracer.Start(ctx, "victron_telemetry",
		trace.WithAttributes(),
	)
	defer span.End()

	if err := w.broker.Publish(ctx, BrokerTopicVictronData, brokerMsg); err != nil {
		slog.Error("publishing victron data to internal broker",
			slog.String("event", eventName),
			slog.Any("error", err),
		)
	}
}

func (w *VictronWorker) Shutdown() {
	slog.Warn("victron worker shutdown not yet implemented")
}

func normalizePath(path string) string {
	b := make([]byte, 0, len(path))
	for i := 0; i < len(path); i++ {
		c := path[i]
		if c >= 'A' && c <= 'Z' {
			b = append(b, '_')
			b = append(b, c+32)
		} else if c == '/' {
			b = append(b, '_')
		} else {
			b = append(b, c)
		}
	}
	return string(b)
}
