package usecases

import (
	"context"
	"fmt"
	"log/slog"

	"zensor-server/cmd/config"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/utils"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricWorker represents a single metric worker that handles one specific metric
type MetricWorker struct {
	config       config.MetricWorkerConfig
	broker       async.InternalBroker
	subscription async.Subscription
	metric       any
	handler      func(ctx context.Context, msg async.BrokerMessage) // Strategy pattern - assigned once at creation
}

// NewMetricWorker creates a new metric worker for a specific metric configuration
func NewMetricWorker(cfg config.MetricWorkerConfig, broker async.InternalBroker) (*MetricWorker, error) {
	meter := otel.Meter("zensor_server")

	var metric any
	var handler func(ctx context.Context, msg async.BrokerMessage)
	var err error

	switch cfg.Type {
	case "counter":
		metric, err = meter.Float64Counter(cfg.Name)
		handler = createCounterHandler(metric, cfg.ValuePropertyName, cfg.CustomAttributes)
	case "gauge":
		metric, err = meter.Float64Gauge(cfg.Name)
		handler = createGaugeHandler(metric, cfg.ValuePropertyName, cfg.CustomAttributes)
	case "histogram":
		metric, err = meter.Float64Histogram(cfg.Name)
		handler = createHistogramHandler(metric, cfg.ValuePropertyName, cfg.CustomAttributes)
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", cfg.Type)
	}

	if err != nil {
		return nil, err
	}

	return &MetricWorker{
		config:  cfg,
		broker:  broker,
		metric:  metric,
		handler: handler,
	}, nil
}

// Run starts the metric worker
func (w *MetricWorker) Run(ctx context.Context, done func()) {
	defer done()

	slog.Info("starting metric worker",
		slog.String("name", w.config.Name),
		slog.String("type", w.config.Type),
		slog.String("topic", w.config.Topic))

	subscription, err := w.broker.Subscribe(async.BrokerTopicName(w.config.Topic))
	if err != nil {
		slog.Error("failed to subscribe to topic",
			slog.String("topic", w.config.Topic),
			slog.Any("error", err))
		return
	}
	defer func() {
		if err := w.broker.Unsubscribe(async.BrokerTopicName(w.config.Topic), subscription); err != nil {
			slog.Error("failed to unsubscribe from topic",
				slog.String("topic", w.config.Topic),
				slog.Any("error", err))
		}
	}()

	w.subscription = subscription

	for {
		select {
		case <-ctx.Done():
			slog.Info("metric worker cancelled", slog.String("name", w.config.Name))
			return
		case msg := <-subscription.Receiver:
			w.handler(ctx, msg)
		}
	}
}

// Shutdown gracefully shuts down the metric worker
func (w *MetricWorker) Shutdown() {
	slog.Info("metric worker shutdown", slog.String("name", w.config.Name))
	if w.subscription.ID != "" {
		if err := w.broker.Unsubscribe(async.BrokerTopicName(w.config.Topic), w.subscription); err != nil {
			slog.Error("failed to unsubscribe during shutdown",
				slog.String("topic", w.config.Topic),
				slog.Any("error", err))
		}
	}
}

func createCounterHandler(metricInstance any, _ string, customAttributes map[string]string) func(context.Context, async.BrokerMessage) {
	return func(ctx context.Context, msg async.BrokerMessage) {
		if counter, ok := metricInstance.(metric.Float64Counter); ok {
			attributes := []attribute.KeyValue{}

			for labelName, path := range customAttributes {
				if value := utils.ExtractStringValue(msg.Value, path); value != "" {
					attributes = append(attributes, attribute.String(labelName, value))
				}
			}

			counter.Add(ctx, 1, metric.WithAttributes(attributes...))
		}
	}
}

func createGaugeHandler(metricInstance any, propertyName string, customAttributes map[string]string) func(ctx context.Context, msg async.BrokerMessage) {
	return func(ctx context.Context, msg async.BrokerMessage) {
		if gauge, ok := metricInstance.(metric.Float64Gauge); ok {
			value := utils.ExtractFloat64Value(msg.Value, propertyName)
			attributes := []attribute.KeyValue{}

			for labelName, path := range customAttributes {
				if value := utils.ExtractStringValue(msg.Value, path); value != "" {
					attributes = append(attributes, attribute.String(labelName, value))
				}
			}

			gauge.Record(ctx, value, metric.WithAttributes(attributes...))
		}
	}
}

func createHistogramHandler(metricInstance any, propertyName string, customAttributes map[string]string) func(context.Context, async.BrokerMessage) {
	return func(ctx context.Context, msg async.BrokerMessage) {
		if histogram, ok := metricInstance.(metric.Float64Histogram); ok {
			value := utils.ExtractFloat64Value(msg.Value, propertyName)
			attributes := []attribute.KeyValue{}

			for labelName, path := range customAttributes {
				if value := utils.ExtractStringValue(msg.Value, path); value != "" {
					attributes = append(attributes, attribute.String(labelName, value))
				}
			}

			histogram.Record(ctx, value, metric.WithAttributes(attributes...))
		}
	}
}
