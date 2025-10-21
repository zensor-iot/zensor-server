package usecases

import (
	"context"
	"log/slog"
	"time"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/shared_kernel/device"
	"zensor-server/internal/shared_kernel/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	_metricKeyTemperature = "temperature"
	_metricKeyHumidity    = "humidity"
	_metricKeyWaterFlow   = "water_flow"
)

func NewMetricPublisherWorker(broker async.InternalBroker) *MetricPublisherWorker {
	return &MetricPublisherWorker{
		broker:         broker,
		metricCounters: make(map[string]metric.Float64Counter),
		metricGauges:   make(map[string]metric.Float64Gauge),
	}
}

var _ async.Worker = &MetricPublisherWorker{}

type MetricPublisherWorker struct {
	broker         async.InternalBroker
	metricCounters map[string]metric.Float64Counter
	metricGauges   map[string]metric.Float64Gauge
}

func (w *MetricPublisherWorker) Run(ctx context.Context, done func()) {
	slog.Debug("metric publisher worker started")
	defer done()

	// Setup metrics
	w.setupMetrics()

	// Subscribe to all relevant topics
	subscriptions := w.setupSubscriptions()
	defer w.cleanupSubscriptions(subscriptions)

	// Process events
	for {
		select {
		case <-ctx.Done():
			slog.Info("metric publisher worker cancelled")
			return
		default:
			// Process all subscription events
			for _, sub := range subscriptions {
				select {
				case msg := <-sub.Receiver:
					w.handleEvent(ctx, msg)
				default:
					// No message available, continue
				}
			}
			time.Sleep(10 * time.Millisecond) // Small delay to prevent busy waiting
		}
	}
}

func (w *MetricPublisherWorker) Shutdown() {
	slog.Info("metric publisher worker shutdown")
}

func (w *MetricPublisherWorker) setupMetrics() {
	meter := otel.Meter("zensor_server")

	// Command metrics
	commandCounter, _ := meter.Float64Counter(
		"zensor_server.commands.total",
		metric.WithDescription("Total number of commands processed"),
	)
	w.metricCounters["commands"] = commandCounter

	// Scheduled task metrics
	scheduledTaskCounter, _ := meter.Float64Counter(
		"zensor_server.scheduled_tasks.total",
		metric.WithDescription("Total number of scheduled tasks executed"),
	)
	w.metricCounters["scheduled_tasks"] = scheduledTaskCounter

	// Sensor metrics (gauges for current values)
	temperatureGauge, _ := meter.Float64Gauge(
		"zensor_server.sensor.temperature",
		metric.WithDescription("Current temperature sensor readings"),
	)
	w.metricGauges[_metricKeyTemperature] = temperatureGauge

	humidityGauge, _ := meter.Float64Gauge(
		"zensor_server.sensor.humidity",
		metric.WithDescription("Current humidity sensor readings"),
	)
	w.metricGauges[_metricKeyHumidity] = humidityGauge

	waterFlowGauge, _ := meter.Float64Gauge(
		"zensor_server.sensor.water_flow",
		metric.WithDescription("Current water flow sensor readings"),
	)
	w.metricGauges[_metricKeyWaterFlow] = waterFlowGauge

	slog.Info("metric publisher worker metrics initialized")
}

func (w *MetricPublisherWorker) setupSubscriptions() []async.Subscription {
	var subscriptions []async.Subscription

	// Subscribe to device messages topic (for sensor data)
	deviceMessagesSub, err := w.broker.Subscribe(async.BrokerTopicName("device_messages"))
	if err != nil {
		slog.Error("failed to subscribe to device_messages topic", slog.Any("error", err))
	} else {
		subscriptions = append(subscriptions, deviceMessagesSub)
	}

	// Subscribe to command events topic
	commandEventsSub, err := w.broker.Subscribe(async.BrokerTopicName("command_events"))
	if err != nil {
		slog.Error("failed to subscribe to command_events topic", slog.Any("error", err))
	} else {
		subscriptions = append(subscriptions, commandEventsSub)
	}

	// Subscribe to scheduled task events topic
	scheduledTaskEventsSub, err := w.broker.Subscribe(async.BrokerTopicName("scheduled_task_events"))
	if err != nil {
		slog.Error("failed to subscribe to scheduled_task_events topic", slog.Any("error", err))
	} else {
		subscriptions = append(subscriptions, scheduledTaskEventsSub)
	}

	return subscriptions
}

func (w *MetricPublisherWorker) cleanupSubscriptions(subscriptions []async.Subscription) {
	for _, sub := range subscriptions {
		if err := w.broker.Unsubscribe(async.BrokerTopicName("device_messages"), sub); err != nil {
			slog.Error("failed to unsubscribe", slog.Any("error", err))
		}
	}
}

func (w *MetricPublisherWorker) handleEvent(ctx context.Context, msg async.BrokerMessage) {
	switch msg.Event {
	case "command_sent":
		w.handleCommandSent(ctx, msg.Value.(device.Command))
	case "command_processed":
		w.handleCommandProcessed(ctx, msg.Value.(device.Command))
	case "command_status_update":
		w.handleCommandStatusUpdate(ctx, msg.Value.(domain.CommandStatusUpdate))
	case "scheduled_task_executed":
		w.handleScheduledTaskExecuted(ctx, msg.Value)
	case "uplink":
		w.handleSensorDataReceived(ctx, msg.Value)
	default:
		slog.Debug("unhandled event type", slog.String("event", msg.Event))
	}
}

func (w *MetricPublisherWorker) handleCommandSent(ctx context.Context, cmd device.Command) {
	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String("zensor_server"),
		attribute.String("device_name", cmd.DeviceName),
		attribute.String("device_id", cmd.DeviceID),
		attribute.String("task_id", cmd.TaskID),
	}

	w.metricCounters["commands"].Add(ctx, 1, metric.WithAttributes(attributes...))
	slog.Debug("published command metric", slog.String("device", cmd.DeviceName))
}

func (w *MetricPublisherWorker) handleCommandProcessed(ctx context.Context, cmd device.Command) {
	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String("zensor_server"),
		attribute.String("device_name", cmd.DeviceName),
		attribute.String("device_id", cmd.DeviceID),
		attribute.String("task_id", cmd.TaskID),
	}

	w.metricCounters["commands"].Add(ctx, 1, metric.WithAttributes(attributes...))
	slog.Debug("published command processed metric", slog.String("device", cmd.DeviceName))
}

func (w *MetricPublisherWorker) handleCommandStatusUpdate(ctx context.Context, statusUpdate domain.CommandStatusUpdate) {
	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String("zensor_server"),
		attribute.String("device_name", statusUpdate.DeviceName),
		attribute.String("status", string(statusUpdate.Status)),
	}

	// Increment counter for status updates
	w.metricCounters["commands"].Add(ctx, 1, metric.WithAttributes(attributes...))
	slog.Debug("published command status update metric", slog.String("device", statusUpdate.DeviceName))
}

func (w *MetricPublisherWorker) handleScheduledTaskExecuted(ctx context.Context, data any) {
	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String("zensor_server"),
	}

	w.metricCounters["scheduled_tasks"].Add(ctx, 1, metric.WithAttributes(attributes...))
	slog.Debug("published scheduled task metric")
}

func (w *MetricPublisherWorker) handleSensorDataReceived(ctx context.Context, data any) {
	// This would handle sensor data and update gauges
	// Implementation depends on the data structure
	slog.Debug("handling sensor data for metrics")
}
