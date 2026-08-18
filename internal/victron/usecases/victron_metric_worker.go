// Package usecases provides business logic use cases for the Victron integration.
package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"zensor-server/internal/infra/async"

	victrondto "zensor-server/internal/victron/dto"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	victronMetricPrefix = "zensor_server_victron"
	victronMetricMeter  = "victron_metric_worker"
)

// VictronMetricWorker records every numeric Victron telemetry point arriving
// through the internal broker as an OTel float64 gauge, one instrument per
// service type and path (e.g. zensor_server_victron_battery_dc_0_voltage).
type VictronMetricWorker struct {
	broker       async.InternalBroker
	subscription async.Subscription
	meter        metric.Meter

	gaugesMu sync.Mutex
	gauges   map[string]metric.Float64Gauge
}

func NewVictronMetricWorker(broker async.InternalBroker) *VictronMetricWorker {
	return &VictronMetricWorker{
		broker: broker,
		meter:  otel.Meter(victronMetricMeter),
		gauges: make(map[string]metric.Float64Gauge),
	}
}

var _ async.Worker = (*VictronMetricWorker)(nil)

func (w *VictronMetricWorker) Run(ctx context.Context, done func()) {
	defer done()

	subscription, err := w.broker.Subscribe(BrokerTopicVictronData)
	if err != nil {
		slog.Error("subscribing to victron metric topic", slog.Any("error", err))
		return
	}
	w.subscription = subscription
	defer func() {
		if err := w.broker.Unsubscribe(BrokerTopicVictronData, subscription); err != nil {
			slog.Error("unsubscribing from victron metric topic", slog.Any("error", err))
		}
	}()

	slog.Info("starting victron metric worker",
		slog.String("topic", string(BrokerTopicVictronData)),
	)

	for {
		select {
		case <-ctx.Done():
			slog.Info("victron metric worker cancelled")
			return
		case msg := <-subscription.Receiver:
			w.handleMessage(msg)
		}
	}
}

func (w *VictronMetricWorker) Shutdown() {
	if w.subscription.ID != "" {
		if err := w.broker.Unsubscribe(BrokerTopicVictronData, w.subscription); err != nil {
			slog.Error("unsubscribing from victron metric topic during shutdown",
				slog.Any("error", err))
		}
	}
}

func (w *VictronMetricWorker) handleMessage(msg async.BrokerMessage) {
	telemetry, ok := msg.Value.(victrondto.VictronTelemetry)
	if !ok {
		return
	}
	if !telemetry.Value.IsNumeric() {
		return
	}

	name := victronMetricName(telemetry)
	gauge, err := w.gauge(name)
	if err != nil {
		slog.Error("creating victron metric gauge",
			slog.String("name", name),
			slog.Any("error", err))
		return
	}

	gauge.Record(context.Background(), telemetry.Value.Value,
		metric.WithAttributes(victronMetricAttributes(telemetry)...),
	)
}

func (w *VictronMetricWorker) gauge(name string) (metric.Float64Gauge, error) {
	w.gaugesMu.Lock()
	defer w.gaugesMu.Unlock()

	if gauge, ok := w.gauges[name]; ok {
		return gauge, nil
	}

	gauge, err := w.meter.Float64Gauge(name)
	if err != nil {
		return nil, err
	}
	w.gauges[name] = gauge
	return gauge, nil
}

func victronMetricName(telemetry victrondto.VictronTelemetry) string {
	path := normalizePath(telemetry.Path)
	path = strings.Trim(path, "_")
	for strings.Contains(path, "__") {
		path = strings.ReplaceAll(path, "__", "_")
	}
	return fmt.Sprintf("%s_%s_%s",
		victronMetricPrefix,
		victrondto.ToSnakeName(telemetry.ServiceType),
		path,
	)
}

func victronMetricAttributes(telemetry victrondto.VictronTelemetry) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("portal_id", telemetry.PortalID),
		attribute.String("service_type", telemetry.ServiceType),
		attribute.Int("instance", telemetry.Instance),
		attribute.String("path", telemetry.Path),
	}
}
