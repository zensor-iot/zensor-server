package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
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

	// fullPublishRetryInterval paces the initial requests for a full publish.
	// Venus only streams values that change, so until the GX has completed one
	// full publish the snapshot is missing every slow-moving path — the
	// battery state of charge and the AC input state among them.
	fullPublishRetryInterval = time.Second

	fullPublishCompletedTopic = "full_publish_completed"
	heartbeatTopic            = "heartbeat"
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
	portalID      string
	mqttClient    mqtt.Client
	broker        async.InternalBroker
	snapshot      *VictronSystemState
	fullPublished atomic.Bool
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
	w.requestFullPublish(ctx)

	ticker := time.NewTicker(keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("victron worker cancelled")
			return
		case <-ticker.C:
			if err := w.sendKeepAlive(ctx); err != nil {
				slog.Error("publishing victron keepalive", slog.Any("error", err))
			}
		}
	}
}

// requestFullPublish asks the GX to republish every topic and keeps asking
// until it confirms one, since the first attempt is lost whenever the MQTT
// connection is still being established.
func (w *VictronWorker) requestFullPublish(ctx context.Context) {
	retry := time.NewTicker(fullPublishRetryInterval)
	defer retry.Stop()

	for {
		if err := w.sendKeepAlive(ctx); err != nil {
			slog.Debug("requesting victron full publish", slog.Any("error", err))
		}

		select {
		case <-ctx.Done():
			return
		case <-retry.C:
		}

		if w.fullPublished.Load() {
			slog.Info("victron full publish completed",
				slog.String("portal_id", w.portalID),
			)
			return
		}
	}
}

func (w *VictronWorker) sendKeepAlive(ctx context.Context) error {
	topic := fmt.Sprintf("R/%s/keepalive", w.portalID)
	if err := w.mqttClient.Publish(ctx, topic, ""); err != nil {
		return fmt.Errorf("publishing victron keepalive to %s: %w", topic, err)
	}

	return nil
}

func (w *VictronWorker) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	switch {
	case strings.HasSuffix(msg.Topic(), "/"+fullPublishCompletedTopic):
		w.fullPublished.Store(true)
		return
	case strings.HasSuffix(msg.Topic(), "/"+heartbeatTopic):
		return
	}

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
	for i := range len(path) {
		c := path[i]
		switch {
		case c >= 'A' && c <= 'Z':
			b = append(b, '_')
			b = append(b, c+32)
		case c == '/':
			b = append(b, '_')
		default:
			b = append(b, c)
		}
	}
	return string(b)
}
