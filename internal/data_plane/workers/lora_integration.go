package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/mqtt"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/device"
	"zensor-server/internal/shared_kernel/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	BrokerTopicUplinkMessage async.BrokerTopicName = "device_messages"

	_defaultQoS byte = 0 // At most once
)

func NewLoraIntegrationWorker(
	ticker *time.Ticker,
	service usecases.DeviceService,
	stateCache usecases.DeviceStateCacheService,
	mqttClient mqtt.Client,
	broker async.InternalBroker,
	commandRepository usecases.CommandRepository,
) *LoraIntegrationWorker {
	return &LoraIntegrationWorker{
		ticker:            ticker,
		service:           service,
		stateCache:        stateCache,
		mqttClient:        mqttClient,
		broker:            broker,
		commandRepository: commandRepository,
	}
}

var _ async.Worker = &LoraIntegrationWorker{}

type LoraIntegrationWorker struct {
	ticker            *time.Ticker
	service           usecases.DeviceService
	stateCache        usecases.DeviceStateCacheService
	mqttClient        mqtt.Client
	broker            async.InternalBroker
	commandRepository usecases.CommandRepository
	devices           sync.Map
}

func (w *LoraIntegrationWorker) Run(ctx context.Context, done func()) {
	slog.Debug("run with context initialized")
	defer done()
	var wg sync.WaitGroup

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

			wg.Add(1)
			dispatchCtx := context.Background()
			dispatchCtx, _ = otel.Tracer("zensor_server").Start(dispatchCtx, "dispatch_ready_commands")
			go w.dispatchReadyCommands(dispatchCtx, wg.Done)
		}
	}
}

func (w *LoraIntegrationWorker) dispatchReadyCommands(ctx context.Context, done func()) {
	defer done()
	span := trace.SpanFromContext(ctx)

	commands, err := w.commandRepository.FindAllReadyToDispatch(ctx)
	if err != nil {
		slog.Error("finding commands ready to dispatch",
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
			slog.Any("error", err),
		)
		return
	}

	for _, cmd := range commands {
		w.dispatchCommand(ctx, cmd)
	}
}

func (w *LoraIntegrationWorker) reconciliation(ctx context.Context, done func()) {
	slog.Debug("reconciliation start...", slog.Time("time", time.Now()))
	span := trace.SpanFromContext(ctx)

	defer done()
	devices, _, err := w.service.AllDevices(ctx, usecases.Pagination{Limit: 1000, Offset: 0})
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

	topicBase = "v3/my-new-application-2021@ttn/devices"
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
		err := w.mqttClient.Subscribe(topic, _defaultQoS, w.messageHandler(ctx))
		if err != nil {
			slog.Error("failed to subscribe to topic",
				slog.String("topic", topic),
				slog.String("error", err.Error()),
			)
		}
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
		case "down/queued":
			w.handleDownlinkResponse(ctx, msg, domain.CommandStatusQueued)
		case "down/sent":
			w.handleDownlinkResponse(ctx, msg, domain.CommandStatusSent)
		case "down/failed":
			w.handleDownlinkResponse(ctx, msg, domain.CommandStatusFailed)
		case "down/ack":
			w.handleDownlinkResponse(ctx, msg, domain.CommandStatusAck)
		default:
			slog.Warn("topic handler not yet implemented",
				slog.String("topic", topicSuffix),
				slog.String("trace_id", span.SpanContext().TraceID().String()),
				slog.String("span_id", span.SpanContext().SpanID().String()),
			)
		}
	}
}

func (w *LoraIntegrationWorker) handleDownlinkResponse(ctx context.Context, msg mqtt.Message, status domain.CommandStatus) {
	span := trace.SpanFromContext(ctx)
	var envelop dto.Envelop
	err := json.Unmarshal(msg.Payload(), &envelop)
	if err != nil {
		slog.Error("failed to unmarshal downlink response message",
			slog.String("error", err.Error()),
			slog.String("status", string(status)),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		return
	}

	logMessage := fmt.Sprintf("downlink %s", status)
	logAttrs := []any{
		slog.String("topic", msg.Topic()),
		slog.String("status", string(status)),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
		slog.String("correlation_ids", fmt.Sprintf("%v", envelop.CorrelationIDs)),
	}

	switch status {
	case domain.CommandStatusFailed:
		logAttrs = append(logAttrs, slog.String("error", envelop.Error.MessageFormat))
		slog.Error(logMessage, logAttrs...)
	default:
		slog.Debug(logMessage, logAttrs...)
	}

	var errorMessage *string
	if status == domain.CommandStatusFailed {
		errorMessage = &envelop.Error.MessageFormat
	}

	w.updateCommandStatus(ctx, envelop, status, errorMessage)
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

	w.handleSensorData(ctx, envelop)
}

func (w *LoraIntegrationWorker) dispatchCommand(ctx context.Context, cmd domain.Command) {
	span := trace.SpanFromContext(ctx)

	command := domainCommandToDeviceCommand(cmd)

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
			{
				FPort:          command.Port,
				Priority:       command.Priority,
				FrmPayload:     rawPayload,
				CorrelationIDs: []string{"zensor:" + command.ID},
			},
		},
	}
	slog.Debug("ttn message",
		slog.Any("msg", ttnMsg),
	)
	err = w.mqttClient.Publish(ctx, topic, ttnMsg)
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

	if err := w.broker.Publish(
		ctx,
		BrokerTopicUplinkMessage,
		async.BrokerMessage{
			Event: "command_sent",
			Value: *command,
		},
	); err != nil {
		slog.Error("failed to publish command sent event", slog.Any("error", err))
	}
}

func domainCommandToDeviceCommand(cmd domain.Command) *device.Command {
	return &device.Command{
		ID:         cmd.ID.String(),
		Version:    int(cmd.Version),
		DeviceID:   cmd.Device.ID.String(),
		DeviceName: cmd.Device.Name,
		TaskID:     cmd.Task.ID.String(),
		Payload: device.CommandPayload{
			Index: uint8(cmd.Payload.Index),
			Value: uint8(cmd.Payload.Value),
		},
		DispatchAfter: cmd.DispatchAfter,
		Port:          uint8(cmd.Port),
		Priority:      string(cmd.Priority),
		CreatedAt:     cmd.CreatedAt,
		Ready:         cmd.Ready,
		Sent:          cmd.Sent,
		SentAt:        cmd.SentAt,
	}
}

func (w *LoraIntegrationWorker) updateCommandStatus(ctx context.Context, envelop dto.Envelop, status domain.CommandStatus, errorMessage *string) {
	span := trace.SpanFromContext(ctx)

	slog.Info("processing command status update",
		slog.String("device_name", envelop.EndDeviceIDs.DeviceID),
		slog.String("status", string(status)),
		slog.Any("correlation_ids", envelop.CorrelationIDs),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
	)

	if len(envelop.CorrelationIDs) == 0 {
		slog.Warn("no correlation IDs found in TTN response",
			slog.String("device_name", envelop.EndDeviceIDs.DeviceID),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		return
	}

	var commandID string
	for _, correlationID := range envelop.CorrelationIDs {
		if id, found := strings.CutPrefix(correlationID, "zensor:"); found {
			commandID = id
			break
		}
	}

	if commandID == "" && len(envelop.CorrelationIDs) > 0 {
		slog.Warn("no zensor correlation ID found in TTN response",
			slog.String("device_name", envelop.EndDeviceIDs.DeviceID),
			slog.Any("correlation_ids", envelop.CorrelationIDs),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		return
	}

	deviceName := envelop.EndDeviceIDs.DeviceID

	statusUpdate := domain.CommandStatusUpdate{
		CommandID:    commandID,
		DeviceName:   deviceName,
		Status:       status,
		ErrorMessage: errorMessage,
		Timestamp:    time.Now(),
	}

	brokerMsg := async.BrokerMessage{
		Event: "command_status_update",
		Value: statusUpdate,
	}

	err := w.broker.Publish(ctx, BrokerTopicUplinkMessage, brokerMsg)
	if err != nil {
		slog.Error("failed to publish command status update",
			slog.String("command_id", commandID),
			slog.String("device_name", deviceName),
			slog.String("status", string(status)),
			slog.String("error", err.Error()),
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		return
	}

	slog.Info("command status update published successfully",
		slog.String("command_id", commandID),
		slog.String("device_name", deviceName),
		slog.String("status", string(status)),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
	)
}

type SensorDataReceived struct {
	DeviceName string  `json:"device_name"`
	AppID      string  `json:"app_id"`
	Value      float64 `json:"value"`
	Index      uint    `json:"index"`
}

func (w *LoraIntegrationWorker) handleSensorData(ctx context.Context, envelope dto.Envelop) {
	slog.Warn("handleSensorData called")
	deviceID := envelope.EndDeviceIDs.DeviceID
	appID := envelope.EndDeviceIDs.ApplicationIDs["application_id"]
	uplink := envelope.UplinkMessage

	for sensorType, sensorDataArray := range uplink.DecodedPayload {
		for _, sensorData := range sensorDataArray {
			sensorDataReceived := SensorDataReceived{
				DeviceName: deviceID,
				AppID:      appID,
				Value:      sensorData.Value,
				Index:      sensorData.Index,
			}

			eventName := utils.ToSnakeCase(sensorType) + "_data_received"
			brokerMsg := async.BrokerMessage{
				Event: eventName,
				Value: sensorDataReceived,
			}

			if err := w.broker.Publish(ctx, BrokerTopicUplinkMessage, brokerMsg); err != nil {
				slog.Error("failed to publish sensor data to internal broker",
					slog.String("device_id", deviceID),
					slog.String("sensor_type", sensorType),
					slog.Uint64("index", uint64(sensorData.Index)),
					slog.Any("error", err))
				return
			}

			slog.Warn("published sensor data to internal broker",
				slog.String("device_id", deviceID),
				slog.String("sensor_type", sensorType),
				slog.Uint64("index", uint64(sensorData.Index)),
				slog.Float64("value", sensorData.Value))
		}
	}
}

func (w *LoraIntegrationWorker) Shutdown() {
	slog.Error("implement me")
}
