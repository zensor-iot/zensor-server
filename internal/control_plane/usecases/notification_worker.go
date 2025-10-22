package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/notification"
	"zensor-server/internal/shared_kernel/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	_metricKeyNotifications = "notifications"
	_tasksTopic             = "tasks"
)

func NewNotificationWorker(
	ticker *time.Ticker,
	notificationClient notification.NotificationClient,
	deviceService DeviceService,
	tenantConfigurationService TenantConfigurationService,
	broker async.InternalBroker,
) *NotificationWorker {
	return &NotificationWorker{
		ticker:                     ticker,
		notificationClient:         notificationClient,
		deviceService:              deviceService,
		tenantConfigurationService: tenantConfigurationService,
		broker:                     broker,
		metricCounters:             make(map[string]metric.Float64Counter),
	}
}

var _ async.Worker = &NotificationWorker{}

type NotificationWorker struct {
	ticker                     *time.Ticker
	notificationClient         notification.NotificationClient
	deviceService              DeviceService
	tenantConfigurationService TenantConfigurationService
	broker                     async.InternalBroker
	metricCounters             map[string]metric.Float64Counter
}

func (w *NotificationWorker) Run(ctx context.Context, done func()) {
	slog.Debug("notification worker run with context initialized")
	defer done()

	subscription, err := w.broker.Subscribe(async.BrokerTopicName(_tasksTopic))
	if err != nil {
		slog.Error("subscribing to tasks topic", slog.Any("error", err))
		return
	}

	if err := w.initializeMetrics(); err != nil {
		slog.Error("initializing metrics", slog.Any("error", err))
		return
	}

	w.processMessages(ctx, subscription)
}

func (w *NotificationWorker) Shutdown() {
	slog.Debug("notification worker shutdown")
}

func (w *NotificationWorker) processMessages(ctx context.Context, subscription async.Subscription) {
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			slog.Debug("notification worker context done, waiting for all messages to finish processing")
			wg.Wait()
			slog.Debug("all notification messages processed, worker shutting down")
			return
		case msg := <-subscription.Receiver:
			wg.Add(1)
			go w.processTaskMessage(ctx, msg, wg.Done)
		}
	}
}

func (w *NotificationWorker) processTaskMessage(ctx context.Context, message async.BrokerMessage, done func()) {
	ctx, span := otel.Tracer("notification-worker").Start(ctx, "process-task-notification")
	defer span.End()
	defer done()

	taskData, ok := message.Value.(map[string]any)
	if !ok {
		slog.Warn("invalid task message format", slog.Any("value", message.Value))
		span.RecordError(fmt.Errorf("invalid task message format"))
		return
	}

	deviceIDStr, ok := taskData["device_id"].(string)
	if !ok {
		slog.Warn("missing device_id in task message", slog.Any("data", taskData))
		span.RecordError(fmt.Errorf("missing device_id in task message"))
		return
	}

	deviceID := domain.ID(deviceIDStr)
	span.SetAttributes(attribute.String("device.id", deviceIDStr))

	device, err := w.deviceService.GetDevice(ctx, deviceID)
	if err != nil {
		slog.Warn("failed to get device for notification",
			slog.String("device_id", deviceIDStr),
			slog.Any("error", err))
		span.RecordError(fmt.Errorf("failed to get device: %w", err))
		return
	}

	if device.TenantID == nil {
		slog.Warn("device has no tenant assigned, skipping notification",
			slog.String("device_id", deviceIDStr))
		span.SetAttributes(attribute.Bool("notification.skipped", true))
		span.SetAttributes(attribute.String("notification.skip_reason", "no_tenant"))
		return
	}

	tenantID := *device.TenantID
	span.SetAttributes(attribute.String("tenant.id", string(tenantID)))

	tenantConfig, err := w.tenantConfigurationService.GetTenantConfiguration(ctx, domain.Tenant{ID: tenantID})
	if err != nil {
		slog.Warn("failed to get tenant configuration for notification",
			slog.String("tenant_id", string(tenantID)),
			slog.Any("error", err))
		span.RecordError(fmt.Errorf("failed to get tenant configuration: %w", err))
		return
	}

	if tenantConfig.NotificationEmail == "" {
		slog.Warn("tenant has no notification email configured, skipping notification",
			slog.String("tenant_id", string(tenantID)))
		span.SetAttributes(attribute.Bool("notification.skipped", true))
		span.SetAttributes(attribute.String("notification.skip_reason", "no_notification_email"))
		return
	}

	if err := w.sendTaskNotification(ctx, device, taskData, tenantConfig.NotificationEmail); err != nil {
		slog.Error("failed to send task notification",
			slog.String("device_id", deviceIDStr),
			slog.String("tenant_id", string(tenantID)),
			slog.String("notification_email", tenantConfig.NotificationEmail),
			slog.Any("error", err))
		span.RecordError(err)
		return
	}

	w.recordNotificationMetrics(ctx, "success")
	span.SetAttributes(attribute.Bool("notification.sent", true))

	slog.Info("task notification sent successfully",
		slog.String("device_id", deviceIDStr),
		slog.String("tenant_id", string(tenantID)),
		slog.String("notification_email", tenantConfig.NotificationEmail))
}

func (w *NotificationWorker) sendTaskNotification(ctx context.Context, device domain.Device, taskData map[string]any, notificationEmail string) error {
	subject := fmt.Sprintf("New Task Created for Device: %s", device.DisplayName)
	body := w.createTaskNotificationBody(device, taskData)

	emailRequest := notification.EmailRequest{
		To:      notificationEmail,
		Subject: subject,
		Body:    body,
	}

	return w.notificationClient.SendEmail(ctx, emailRequest)
}

func (w *NotificationWorker) createTaskNotificationBody(device domain.Device, taskData map[string]any) string {
	body := fmt.Sprintf("A new task has been created for device: %s\n\n", device.DisplayName)
	body += "Device Details:\n"
	body += "- Device ID: " + string(device.ID) + "\n"
	body += "- Device Name: " + device.Name + "\n"
	body += "- Display Name: " + device.DisplayName + "\n"
	body += "- App EUI: " + device.AppEUI + "\n"
	body += "- Dev EUI: " + device.DevEUI + "\n"

	if device.TenantID != nil {
		body += "- Tenant ID: " + string(*device.TenantID) + "\n"
	}

	body += "\nTask Details:\n"
	body += "- Task ID: " + fmt.Sprintf("%v", taskData["id"]) + "\n"
	body += "- Created At: " + fmt.Sprintf("%v", taskData["created_at"]) + "\n"

	if scheduledTaskID, ok := taskData["scheduled_task_id"]; ok && scheduledTaskID != nil {
		body += "- Scheduled Task ID: " + fmt.Sprintf("%v", scheduledTaskID) + "\n"
	}

	body += "\nThis is an automated notification from Zensor Server.\n"

	return body
}

func (w *NotificationWorker) initializeMetrics() error {
	meter := otel.Meter("notification-worker")

	notificationCounter, err := meter.Float64Counter(
		"zensor_server_notifications_total",
		metric.WithDescription("Total number of notifications sent"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("creating notification counter: %w", err)
	}

	w.metricCounters[_metricKeyNotifications] = notificationCounter
	return nil
}

func (w *NotificationWorker) recordNotificationMetrics(ctx context.Context, status string) {
	if counter, exists := w.metricCounters[_metricKeyNotifications]; exists {
		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("status", status),
			semconv.ServiceNameKey.String("zensor-server"),
			semconv.ServiceVersionKey.String("1.0.0"),
		))
	}
}
