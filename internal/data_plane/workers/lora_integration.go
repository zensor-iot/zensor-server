package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"sync"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/pubsub"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	BrokerTopicUplinkMessage  async.BrokerTopicName = "device_messages"
	PubSubTopicDeviceCommands pubsub.Topic          = "device_commands"
)

func NewLoraIntegrationWorker(
	ticker *time.Ticker,
	service usecases.DeviceService,
	mqttClient mqtt.Client,
	broker async.InternalBroker,
	pubsubConsumerFactory pubsub.ConsumerFactory,
) *LoraIntegrationWorker {
	return &LoraIntegrationWorker{
		ticker:         ticker,
		service:        service,
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
			go w.reconciliation(tickCtx, wg.Done)
		case msg := <-commandsChannel:
			wg.Add(1)
			go w.deviceCommandHandler(msg, wg.Done)
		}
	}
}

const (
	_metricKeyTemperature = "temperature"
	_metricKeyHumidity    = "humidity"
)

func (w *LoraIntegrationWorker) setupOtelCounters() {
	meter := otel.Meter("zensor_server")
	temperatureCounter, _ := meter.Float64Gauge(fmt.Sprintf("%s.%s", "zensor_server", "sensor.temperature"), metric.WithDescription("zensor_server temperature metric counter"))
	humidityCounter, _ := meter.Float64Gauge(fmt.Sprintf("%s.%s", "zensor_server", "sensor.humidity"), metric.WithDescription("zensor_server humidity metric counter"))

	w.metricCounters[_metricKeyTemperature] = temperatureCounter
	w.metricCounters[_metricKeyHumidity] = humidityCounter
}

func (w *LoraIntegrationWorker) consumeCommandsToChannel() <-chan pubsub.Prototype {
	out := make(chan pubsub.Prototype, 1)
	handler := func(msg pubsub.Prototype) error {
		out <- msg
		return nil
	}

	go w.pubsubConsumer.Consume(PubSubTopicDeviceCommands, handler, dto.Command{})
	return out
}

func (w *LoraIntegrationWorker) reconciliation(ctx context.Context, done func()) {
	slog.Debug("reconciliation start...", slog.Time("time", time.Now()))
	defer done()
	devices, err := w.service.AllDevices(ctx)
	if err != nil {
		slog.Error("getting all devices", slog.Any("error", err))
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
	slog.Debug("handle device", slog.Any("device", device))
	if _, exists := w.devices.Load(device.ID); exists {
		slog.Debug("device is already configured", slog.String("device", device.Name))
		return
	}
	w.devices.Store(device.ID, device)
	for _, suffix := range topics {
		topic := fmt.Sprintf("%s/%s/%s", topicBase, device.Name, suffix)
		slog.Debug("final topic", slog.String("value", topic))
		w.mqttClient.Subscribe(topic, qos, w.messageHandler(ctx))
	}
}

var topicRegex = regexp.MustCompile(`^.*/devices/[\w-_]*/(.*)$`)

func (w *LoraIntegrationWorker) messageHandler(ctx context.Context) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		slog.Info("message received",
			slog.String("topic", msg.Topic()),
			slog.Uint64("message_id", uint64(msg.MessageID())),
		)

		result := topicRegex.FindStringSubmatch(msg.Topic())
		if len(result) < 2 {
			slog.Error("invalid topic", slog.String("topic", msg.Topic()))
			return
		}

		topicSuffix := result[1]
		switch topicSuffix {
		case "up":
			w.uplinkMessageHandler(ctx, msg)
		default:
			slog.Warn("topic handler not yet implemented", slog.String("topic", topicSuffix))
		}
	}
}

func (w *LoraIntegrationWorker) uplinkMessageHandler(ctx context.Context, msg mqtt.Message) {
	var envelop dto.Envelop
	json.Unmarshal(msg.Payload(), &envelop)
	decodedPayload := envelop.UplinkMessage.FromMessagePack()
	fmt.Printf("*** %+v\n", decodedPayload)
	brokerMsg := async.BrokerMessage{
		Event: "uplink",
		Value: envelop,
	}
	w.broker.Publish(ctx, BrokerTopicUplinkMessage, brokerMsg)
	slog.Debug("envelop", slog.Any("envelop", envelop))

	var (
		temperature float64 = -1
		humidity    float64 = -1
	)

	if val, ok := envelop.UplinkMessage.DecodedPayload[_metricKeyTemperature].(float64); ok {
		temperature = val
	}

	if val, ok := envelop.UplinkMessage.DecodedPayload[_metricKeyHumidity].(float64); ok {
		humidity = val
	}
	
	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String("zensor_server"),
		attribute.String("device_name", envelop.EndDeviceIDs.DeviceID),
		attribute.String("app_id", envelop.EndDeviceIDs.ApplicationIDs["application_id"]),
	}
	w.metricCounters[_metricKeyTemperature].Record(ctx, temperature, metric.WithAttributes(attributes...))
	w.metricCounters[_metricKeyHumidity].Record(ctx, humidity, metric.WithAttributes(attributes...))
}

func (w *LoraIntegrationWorker) deviceCommandHandler(msg pubsub.Prototype, done func()) {
	defer done()
	command, ok := msg.(*dto.Command)
	if !ok {
		slog.Error("parsing message type", slog.String("error", "msg is not dto.Command"), slog.Any("message", msg))
		return
	}
	topic := fmt.Sprintf("%s/%s/%s", topicBase, command.DeviceName, "down/push")
	rawPayload, err := command.Payload.ToMessagePack()
	if err != nil {
		slog.Error("converting to message pack failed",
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
		slog.Error("publishing command", slog.String("device_id", command.DeviceID), slog.String("error", err.Error()))
		return
	}

	slog.Debug("message published",
		slog.String("device_id", command.DeviceID),
		slog.String("topic", topic),
	)
}

func (w *LoraIntegrationWorker) Shutdown() {
	panic("implement me")
}
