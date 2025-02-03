package workers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/mqtt"
)

func NewLoraIntegrationWorker(
	ticker *time.Ticker,
	service usecases.DeviceService,
	mqttClient mqtt.Client,
) *LoraIntegrationWorker {
	return &LoraIntegrationWorker{
		ticker:     ticker,
		service:    service,
		mqttClient: mqttClient,
	}
}

var _ async.Worker = &LoraIntegrationWorker{}

type LoraIntegrationWorker struct {
	ticker     *time.Ticker
	service    usecases.DeviceService
	mqttClient mqtt.Client
	devices    sync.Map
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
			go w.reconciliation(tickCtx, wg.Done)
		}
	}
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

	topicBase           = "v3/my-new-application-2021@ttn/devices"
	qos            byte = 0
	messageHandler      = func(_ mqtt.Client, msg mqtt.Message) {
		slog.Info("message received",
			slog.String("topic", msg.Topic()),
			slog.Uint64("message_id", uint64(msg.MessageID())),
			slog.String("payload", string(msg.Payload())),
		)
	}
)

func (w *LoraIntegrationWorker) handleDevice(_ context.Context, device domain.Device) {
	slog.Debug("handle device", slog.Any("device", device))
	if _, exists := w.devices.Load(device.ID); exists {
		slog.Debug("device is already configured", slog.String("device", device.Name))
		return
	}
	w.devices.Store(device.ID, device)
	for _, suffix := range topics {
		topic := fmt.Sprintf("%s/%s/%s", topicBase, device.Name, suffix)
		slog.Debug("final topic", slog.String("value", topic))
		w.mqttClient.Subscribe(topic, qos, messageHandler)
	}
}

func (w *LoraIntegrationWorker) Shutdown() {
	panic("implement me")
}
