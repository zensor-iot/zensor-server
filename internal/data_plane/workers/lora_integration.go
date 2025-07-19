package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"sync"
	"time"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/avro"
	"zensor-server/internal/shared_kernel/device"
	"zensor-server/internal/shared_kernel/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	BrokerTopicUplinkMessage  async.BrokerTopicName = "device_messages"
	PubSubTopicDeviceCommands pubsub.Topic          = "device_commands"
)

func NewLoraIntegrationWorker(
	ticker *time.Ticker,
	service usecases.DeviceService,
	stateCache usecases.DeviceStateCacheService,
	mqttClient mqtt.Client,
	broker async.InternalBroker,
	pubsubConsumerFactory pubsub.ConsumerFactory,
) *LoraIntegrationWorker {
	return &LoraIntegrationWorker{
		ticker:         ticker,
		service:        service,
		stateCache:     stateCache,
		mqttClient:     mqttClient,
		broker:         broker,
		pubsubConsumer: pubsubConsumerFactory.New(),
		metricCounters: make(map[string]metric.Float64Gauge),
	}
}

var _ async.Worker = &LoraIntegrationWorker{}

type LoraIntegrationWorker struct {
	ticker         *time.Ticker
	service        usecases.DeviceService
	stateCache     usecases.DeviceStateCacheService
	mqttClient     mqtt.Client
	broker         async.InternalBroker
	pubsubConsumer pubsub.Consumer
	devices        sync.Map
	metricCounters map[string]metric.Float64Gauge
}

func (w *LoraIntegrationWorker) Run(ctx context.Context, done func()) {
	slog.Debug("run with context initialized")
	defer done()
	var wg sync.WaitGroup
	commandsChannel := w.consumeCommandsToChannel()
	w.setupOtelCounters()
	for {
		select {
		case <-ctx.Done():
			slog.Warn("lora intetration worker cancelled")
			wg.Wait()
			return
		case <-w.ticker.C:
			wg.Add(1)
			tickCtx := context.Background()
			tickCtx, _ = otel.Tracer("zensor_server").Start(tickCtx, "device_command_handler")
			go w.reconciliation(tickCtx, wg.Done)
		case msg := <-commandsChannel:
			wg.Add(1)
			span := trace.SpanFromContext(msg.Ctx)
			procCtx := context.Background()
			if span.SpanContext().HasSpanID() {
				procCtx = trace.ContextWithSpan(procCtx, span)
			} else {
				procCtx, _ = otel.Tracer("zensor_server").Start(procCtx, "device_command_handler")
			}
			go w.deviceCommandHandler(procCtx, msg, wg.Done)
		}
	}
}

const (
	_metricKeyTemperature = "temperature"
	_metricKeyHumidity    = "humidity"
	_metricKeyWaterFlow   = "waterFlow"
)

func (w *LoraIntegrationWorker) setupOtelCounters() {
	meter := otel.Meter("zensor_server")
	temperatureCounter, _ := meter.Float64Gauge(fmt.Sprintf("%s.%s", "zensor_server", "sensor.temperature"), metric.WithDescription("zensor_server temperature metric counter"))
	humidityCounter, _ := meter.Float64Gauge(fmt.Sprintf("%s.%s", "zensor_server", "sensor.humidity"), metric.WithDescription("zensor_server humidity metric counter"))
	waterFlowGauge, _ := meter.Float64Gauge(fmt.Sprintf("%s.%s", "zensor_server", "sensor.water_flow"), metric.WithDescription("zensor_server water flow metric gauge"))

	w.metricCounters[_metricKeyTemperature] = temperatureCounter
	w.metricCounters[_metricKeyHumidity] = humidityCounter
	w.metricCounters[_metricKeyWaterFlow] = waterFlowGauge
}

func (w *LoraIntegrationWorker) consumeCommandsToChannel() <-chan pubsub.ConsumedMessage {
	out := make(chan pubsub.ConsumedMessage, 1)
	handler := func(ctx context.Context, key pubsub.Key, msg pubsub.Prototype) error {
		out <- pubsub.ConsumedMessage{
			Ctx:   ctx,
			Key:   key,
			Value: msg,
		}
		return nil
	}

	go func() {
		err := w.pubsubConsumer.Consume(PubSubTopicDeviceCommands, handler, &avro.AvroCommand{})
		if err != nil {
			slog.Error("failed to consume commands", slog.String("error", err.Error()))
		}
	}()
	return out
}

func (w *LoraIntegrationWorker) reconciliation(ctx context.Context, done func()) {
	slog.Debug("reconciliation start...", slog.Time("time", time.Now()))
	span := trace.SpanFromContext(ctx)

	defer done()
	devices, err := w.service.AllDevices(ctx)
	if err != nil {
		slog.Error("getting all devices",
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
			slog.Any("error", err),
		)
		return
	}

	for _, device := range devices {
		w.handleDevice(ctx, device)
	}
	slog.Debug("reconciliation end", slog.Time("time", time.Now()))
}

var (
	topics = []string{
		"join",
		"up",
		"down/queued",
		"down/sent",
		"down/failed",
		"down/ack",
	}

	topicBase      = "v3/my-new-application-2021@ttn/devices"
	qos       byte = 0
)

func (w *LoraIntegrationWorker) handleDevice(ctx context.Context, device domain.Device) {
	span := trace.SpanFromContext(ctx)
	slog.Debug("handle device",
		slog.String("device", device.Name),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
	)
	if _, exists := w.devices.Load(device.ID); exists {
		slog.Debug("device is already configured",
			slog.String("device", device.Name),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		return
	}
	w.devices.Store(device.ID, device)
	for _, suffix := range topics {
		topic := fmt.Sprintf("%s/%s/%s", topicBase, device.Name, suffix)
		slog.Debug("final topic",
			slog.String("value", topic),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		w.mqttClient.Subscribe(topic, qos, w.messageHandler(ctx))
	}
}

var topicRegex = regexp.MustCompile(`^.*/devices/[\w-_]*/(.*)$`)

func (w *LoraIntegrationWorker) messageHandler(ctx context.Context) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		span := trace.SpanFromContext(ctx)
		slog.Info("message received",
			slog.String("topic", msg.Topic()),
			slog.Uint64("message_id", uint64(msg.MessageID())),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)

		result := topicRegex.FindStringSubmatch(msg.Topic())
		if len(result) < 2 {
			slog.Error("invalid topic",
				slog.String("topic", msg.Topic()),
				slog.String("trace_id", span.SpanContext().TraceID().String()),
				slog.String("span_id", span.SpanContext().SpanID().String()),
			)
			return
		}

		topicSuffix := result[1]
		switch topicSuffix {
		case "up":
			w.uplinkMessageHandler(ctx, msg)
		case "down/failed":
			w.downlinkFailedHandler(ctx, msg)
		default:
			slog.Warn("topic handler not yet implemented",
				slog.String("topic", topicSuffix),
				slog.String("trace_id", span.SpanContext().TraceID().String()),
				slog.String("span_id", span.SpanContext().SpanID().String()),
			)
		}
	}
}

func (w *LoraIntegrationWorker) downlinkFailedHandler(ctx context.Context, msg mqtt.Message) {
	span := trace.SpanFromContext(ctx)
	var envelop dto.Envelop
	err := json.Unmarshal(msg.Payload(), &envelop)
	if err != nil {
		slog.Error("failed to unmarshal message",
			slog.String("error", err.Error()),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		return
	}

	slog.Error("downlink failed",
		slog.String("topic", msg.Topic()),
		slog.String("error", envelop.Error.MessageFormat),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
	)
}

func (w *LoraIntegrationWorker) uplinkMessageHandler(ctx context.Context, msg mqtt.Message) {
	span := trace.SpanFromContext(ctx)
	var envelop dto.Envelop
	err := json.Unmarshal(msg.Payload(), &envelop)
	if err != nil {
		slog.Error("failed to unmarshal message",
			slog.String("error", err.Error()),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		return
	}

	envelop.UplinkMessage.FromMessagePack()

	deviceName := envelop.EndDeviceIDs.DeviceID
	err = w.service.UpdateLastMessageReceivedAt(ctx, deviceName)
	if err != nil {
		slog.Error("failed to update device last message timestamp",
			slog.String("device_name", deviceName),
			slog.String("error", err.Error()),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	// Update device state cache with new sensor data
	err = w.stateCache.SetState(ctx, deviceName, envelop.UplinkMessage.DecodedPayload)
	if err != nil {
		slog.Error("failed to update device state cache",
			slog.String("device_name", deviceName),
			slog.String("error", err.Error()),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	brokerMsg := async.BrokerMessage{
		Event: "uplink",
		Value: envelop,
	}
	err = w.broker.Publish(ctx, BrokerTopicUplinkMessage, brokerMsg)
	if err != nil {
		slog.Error("failed to publish message", slog.String("error", err.Error()))
	}
	slog.Debug("envelop", slog.Any("envelop", envelop))

	temperatures := envelop.UplinkMessage.DecodedPayload[_metricKeyTemperature]
	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String("zensor_server"),
		attribute.String("device_name", envelop.EndDeviceIDs.DeviceID),
		attribute.String("app_id", envelop.EndDeviceIDs.ApplicationIDs["application_id"]),
	}
	for _, t := range temperatures {
		extAttributes := append(attributes, attribute.Int("index", int(t.Index)))
		w.metricCounters[_metricKeyTemperature].Record(ctx, t.Value, metric.WithAttributes(extAttributes...))
	}

	humidities := envelop.UplinkMessage.DecodedPayload[_metricKeyHumidity]
	for _, h := range humidities {
		extAttributes := append(attributes, attribute.Int("index", int(h.Index)))
		w.metricCounters[_metricKeyHumidity].Record(ctx, h.Value, metric.WithAttributes(extAttributes...))
	}

	waterFlows := envelop.UplinkMessage.DecodedPayload[_metricKeyWaterFlow]
	for _, wf := range waterFlows {
		extAttributes := append(attributes, attribute.Int("index", int(wf.Index)))
		w.metricCounters[_metricKeyWaterFlow].Record(ctx, wf.Value, metric.WithAttributes(extAttributes...))
	}
}

func (w *LoraIntegrationWorker) deviceCommandHandler(ctx context.Context, msg pubsub.ConsumedMessage, done func()) {
	defer done()
	span := trace.SpanFromContext(ctx)

	command, err := w.convertToSharedCommand(ctx, msg.Value)
	if err != nil {
		slog.Error("converting message to command",
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
			slog.String("error", err.Error()),
		)
		return
	}

	if !command.Ready {
		slog.Warn("command won't be send because is not ready",
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		return
	}
	if command.Sent {
		slog.Warn("command won't be send because is was already sent",
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		return
	}

	topic := fmt.Sprintf("%s/%s/%s", topicBase, command.DeviceName, "down/push")
	rawPayload, err := command.Payload.ToMessagePack()
	if err != nil {
		slog.Error("converting to message pack failed",
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
			slog.String("error", err.Error()),
		)
		return
	}
	ttnMsg := dto.TTNMessage{
		Downlinks: []dto.TTNMessageDownlink{
			{FPort: command.Port, Priority: command.Priority, FrmPayload: rawPayload},
		},
	}
	slog.Debug("ttn message",
		slog.Any("msg", ttnMsg),
	)
	err = w.mqttClient.Publish(topic, ttnMsg)
	if err != nil {
		slog.Error("publishing command",
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
			slog.String("device_id", command.DeviceID),
			slog.String("error", err.Error()),
		)
		return
	}

	slog.Debug("message published",
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("device_id", command.DeviceID),
		slog.String("topic", topic),
	)

	w.broker.Publish(
		ctx,
		BrokerTopicUplinkMessage,
		async.BrokerMessage{
			Event: "command_sent",
			Value: *command,
		},
	)
}

func (w *LoraIntegrationWorker) convertToSharedCommand(ctx context.Context, msg pubsub.Prototype) (*device.Command, error) {
	span := trace.SpanFromContext(ctx)
	slog.Info("converting to shared command",
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("type", fmt.Sprintf("%T", msg)),
	)

	// Try to convert directly from AvroCommand
	if avroCmd, ok := msg.(*avro.AvroCommand); ok {
		return &device.Command{
			ID:         avroCmd.ID,
			Version:    avroCmd.Version,
			DeviceID:   avroCmd.DeviceID,
			DeviceName: avroCmd.DeviceName,
			TaskID:     avroCmd.TaskID,
			Payload: device.CommandPayload{
				Index: uint8(avroCmd.PayloadIndex),
				Value: uint8(avroCmd.PayloadValue),
			},
			DispatchAfter: utils.Time{Time: avroCmd.DispatchAfter},
			Port:          uint8(avroCmd.Port),
			Priority:      avroCmd.Priority,
			CreatedAt:     utils.Time{Time: avroCmd.CreatedAt},
			Ready:         avroCmd.Ready,
			Sent:          avroCmd.Sent,
			SentAt:        utils.Time{Time: avroCmd.SentAt},
		}, nil
	}

	// Fallback to JSON marshaling/unmarshaling for other types
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshaling message to JSON: %w", err)
	}

	var command device.Command
	err = json.Unmarshal(jsonData, &command)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling to device.Command: %w", err)
	}

	return &command, nil
}

func (w *LoraIntegrationWorker) Shutdown() {
	slog.Error("implement me")
}
