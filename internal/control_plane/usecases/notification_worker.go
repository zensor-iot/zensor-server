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
	_recentTasksLimit       = 10
)

func NewNotificationWorker(
	ticker *time.Ticker,
	notificationClient notification.NotificationClient,
	deviceService DeviceService,
	tenantConfigurationService TenantConfigurationService,
	taskService TaskService,
	broker async.InternalBroker,
) *NotificationWorker {
	return &NotificationWorker{
		ticker:                     ticker,
		notificationClient:         notificationClient,
		deviceService:              deviceService,
		tenantConfigurationService: tenantConfigurationService,
		taskService:                taskService,
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
	taskService                TaskService
	broker                     async.InternalBroker
	metricCounters             map[string]metric.Float64Counter
}

func (w *NotificationWorker) Run(ctx context.Context, done func()) {
	slog.Debug("notification worker run with context initialized")
	defer done()

	subscription, err := w.broker.Subscribe(async.BrokerTopicName(_scheduledTasksTopic))
	if err != nil {
		slog.Error("subscribing to scheduled_tasks topic", slog.Any("error", err))
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
			go w.processScheduledTaskEvent(ctx, msg, wg.Done)
		}
	}
}

func (w *NotificationWorker) processScheduledTaskEvent(ctx context.Context, message async.BrokerMessage, done func()) {
	ctx, span := otel.Tracer("notification-worker").Start(ctx, "process-scheduled-task-event")
	defer span.End()
	defer done()

	if message.Event != "scheduled_task_executed" {
		slog.Debug("ignoring event", slog.String("event", message.Event))
		return
	}

	scheduledTask, ok := message.Value.(domain.ScheduledTask)
	if !ok {
		slog.Warn("invalid scheduled task message format", slog.Any("value", message.Value))
		span.RecordError(fmt.Errorf("invalid scheduled task message format"))
		return
	}

	scheduledTaskID := scheduledTask.ID
	span.SetAttributes(attribute.String("scheduled_task.id", scheduledTaskID.String()))
	span.SetAttributes(attribute.String("device.id", scheduledTask.Device.ID.String()))

	tasks, _, err := w.taskService.FindAllByScheduledTask(ctx, scheduledTaskID, Pagination{
		Limit:  _recentTasksLimit,
		Offset: 0,
	})
	if err != nil {
		slog.Error("failed to find tasks for scheduled task",
			slog.String("scheduled_task_id", scheduledTaskID.String()),
			slog.Any("error", err))
		span.RecordError(fmt.Errorf("failed to find tasks: %w", err))
		return
	}

	if len(tasks) == 0 {
		slog.Debug("no tasks found for scheduled task",
			slog.String("scheduled_task_id", scheduledTaskID.String()))
		return
	}

	for _, task := range tasks {
		w.processTaskNotification(ctx, task, scheduledTask)
	}
}

func (w *NotificationWorker) processTaskNotification(ctx context.Context, task domain.Task, scheduledTask domain.ScheduledTask) {
	ctx, span := otel.Tracer("notification-worker").Start(ctx, "process-task-notification")
	defer span.End()

	deviceID := task.Device.ID
	span.SetAttributes(attribute.String("task.id", task.ID.String()))
	span.SetAttributes(attribute.String("device.id", deviceID.String()))

	device, err := w.deviceService.GetDevice(ctx, deviceID)
	if err != nil {
		slog.Warn("failed to get device for notification",
			slog.String("device_id", deviceID.String()),
			slog.String("task_id", task.ID.String()),
			slog.Any("error", err))
		span.RecordError(fmt.Errorf("failed to get device: %w", err))
		return
	}

	if device.TenantID == nil {
		slog.Warn("device has no tenant assigned, skipping notification",
			slog.String("device_id", deviceID.String()),
			slog.String("task_id", task.ID.String()))
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
			slog.String("task_id", task.ID.String()),
			slog.Any("error", err))
		span.RecordError(fmt.Errorf("failed to get tenant configuration: %w", err))
		return
	}

	if tenantConfig.NotificationEmail == "" {
		slog.Warn("tenant has no notification email configured, skipping notification",
			slog.String("tenant_id", string(tenantID)),
			slog.String("task_id", task.ID.String()))
		span.SetAttributes(attribute.Bool("notification.skipped", true))
		span.SetAttributes(attribute.String("notification.skip_reason", "no_notification_email"))
		return
	}

	if err := w.sendTaskNotification(ctx, device, task, scheduledTask, tenantConfig.NotificationEmail); err != nil {
		slog.Error("failed to send task notification",
			slog.String("device_id", deviceID.String()),
			slog.String("task_id", task.ID.String()),
			slog.String("tenant_id", string(tenantID)),
			slog.String("notification_email", tenantConfig.NotificationEmail),
			slog.Any("error", err))
		span.RecordError(err)
		return
	}

	w.recordNotificationMetrics(ctx, "success")
	span.SetAttributes(attribute.Bool("notification.sent", true))

	slog.Info("task notification sent successfully",
		slog.String("device_id", deviceID.String()),
		slog.String("task_id", task.ID.String()),
		slog.String("scheduled_task_id", scheduledTask.ID.String()),
		slog.String("tenant_id", string(tenantID)),
		slog.String("notification_email", tenantConfig.NotificationEmail))
}

func (w *NotificationWorker) sendTaskNotification(ctx context.Context, device domain.Device, task domain.Task, scheduledTask domain.ScheduledTask, notificationEmail string) error {
	subject := fmt.Sprintf("New Task Created for Device: %s", device.DisplayName)
	body := w.createTaskNotificationBody(device, task, scheduledTask)

	emailRequest := notification.EmailRequest{
		To:      notificationEmail,
		Subject: subject,
		Body:    body,
	}

	return w.notificationClient.SendEmail(ctx, emailRequest)
}

func (w *NotificationWorker) createTaskNotificationBody(device domain.Device, task domain.Task, scheduledTask domain.ScheduledTask) string {
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
	body += "- Task ID: " + task.ID.String() + "\n"
	body += "- Created At: " + task.CreatedAt.Time.Format(time.RFC3339) + "\n"
	body += "- Scheduled Task ID: " + scheduledTask.ID.String() + "\n"
	body += "- Number of Commands: " + fmt.Sprintf("%d", len(task.Commands)) + "\n"

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
