package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/config"
	"zensor-server/internal/infra/notification"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	_metricKeyPushNotifications = "push_notifications"
)

type PushNotificationWorker struct {
	config             config.PushNotificationWorkerConfig
	broker             async.InternalBroker
	subscription       async.Subscription
	notificationClient notification.NotificationClient
	pushTokenService   PushTokenService
	userService        UserService
	metricCounters     map[string]metric.Float64Counter
}

func NewPushNotificationWorker(
	cfg config.PushNotificationWorkerConfig,
	broker async.InternalBroker,
	notificationClient notification.NotificationClient,
	pushTokenService PushTokenService,
	userService UserService,
) (*PushNotificationWorker, error) {
	worker := &PushNotificationWorker{
		config:             cfg,
		broker:             broker,
		notificationClient: notificationClient,
		pushTokenService:   pushTokenService,
		userService:        userService,
		metricCounters:     make(map[string]metric.Float64Counter),
	}

	if err := worker.initializeMetrics(); err != nil {
		return nil, fmt.Errorf("initializing metrics: %w", err)
	}

	worker.warnOnUnsupportedTemplates()

	return worker, nil
}

// warnOnUnsupportedTemplates reports a template still written with the %s
// placeholder that this worker used to support. Only {{key}} is interpolated
// now, so a stale template would reach the user as a literal "%s"; saying so at
// startup turns a silent wrong notification into a visible misconfiguration.
func (w *PushNotificationWorker) warnOnUnsupportedTemplates() {
	templates := map[string]string{
		"title_template":    w.config.TitleTemplate,
		"body_template":     w.config.BodyTemplate,
		"deeplink_template": w.config.DeepLinkTemplate,
	}

	for field, template := range templates {
		if strings.Contains(template, "%s") {
			slog.Warn("push notification template uses an unsupported %s placeholder, use {{key}} instead",
				slog.String("notification", w.config.Name),
				slog.String("field", field),
				slog.String("template", template))
		}
	}
}

var _ async.Worker = (*PushNotificationWorker)(nil)

func (w *PushNotificationWorker) Run(ctx context.Context, done func()) {
	defer done()

	slog.Info("starting push notification worker",
		slog.String("name", w.config.Name),
		slog.String("topic", w.config.Topic),
		slog.String("event_type", w.config.EventType))

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
			slog.Info("push notification worker cancelled", slog.String("name", w.config.Name))
			return
		case msg := <-subscription.Receiver:
			if msg.Event == w.config.EventType {
				w.handleNotification(ctx, msg)
			}
		}
	}
}

func (w *PushNotificationWorker) Shutdown() {
	slog.Info("push notification worker shutdown", slog.String("name", w.config.Name))
	if w.subscription.ID != "" {
		if err := w.broker.Unsubscribe(async.BrokerTopicName(w.config.Topic), w.subscription); err != nil {
			slog.Error("failed to unsubscribe during shutdown",
				slog.String("topic", w.config.Topic),
				slog.Any("error", err))
		}
	}
}

func (w *PushNotificationWorker) handleNotification(ctx context.Context, msg async.BrokerMessage) {
	ctx, span := otel.Tracer("push-notification-worker").Start(ctx, "handle-notification")
	defer span.End()

	span.SetAttributes(attribute.String("notification.name", w.config.Name))
	span.SetAttributes(attribute.String("event.type", msg.Event))

	tenantIDStr := utils.ExtractStringValue(msg.Value, w.config.TenantIDPath)
	if tenantIDStr == "" {
		slog.Warn("tenant ID not found in message",
			slog.String("path", w.config.TenantIDPath),
			slog.String("notification", w.config.Name))
		span.SetAttributes(attribute.Bool("notification.skipped", true))
		span.SetAttributes(attribute.String("notification.skip_reason", "no_tenant_id"))
		return
	}

	tenantID := domain.ID(tenantIDStr)
	span.SetAttributes(attribute.String("tenant.id", tenantIDStr))

	userIDStr := utils.ExtractStringValue(msg.Value, w.config.UserIDPath)
	if userIDStr == "" {
		slog.Warn("user ID not found in message, will send to all tenant users",
			slog.String("path", w.config.UserIDPath),
			slog.String("notification", w.config.Name))
		w.sendToTenantUsers(ctx, tenantID, msg)
		return
	}

	userID := domain.ID(userIDStr)
	w.sendToUser(ctx, userID, msg)
}

func (w *PushNotificationWorker) sendToUser(ctx context.Context, userID domain.ID, msg async.BrokerMessage) {
	tokens, err := w.pushTokenService.ListTokensByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrPushTokenNotFound) {
			slog.Debug("user has no push token registered",
				slog.String("user_id", userID.String()),
				slog.String("notification", w.config.Name))
			return
		}
		slog.Warn("failed to list push tokens for user",
			slog.String("user_id", userID.String()),
			slog.Any("error", err))
		return
	}

	title := w.buildTitle(msg)
	body := w.buildBody(msg)
	deepLink := w.buildDeepLink(msg)

	anySuccess := false
	for _, pushToken := range tokens {
		request := notification.PushNotificationRequest{
			Token:    pushToken.Token,
			Title:    title,
			Body:     body,
			DeepLink: deepLink,
			Platform: pushToken.Platform,
		}

		err = w.notificationClient.SendPushNotification(ctx, request)
		if err != nil {
			slog.Error("failed to send push notification",
				slog.String("user_id", userID.String()),
				slog.String("notification", w.config.Name),
				slog.Any("error", err))
			w.recordNotificationMetrics(ctx, "error")
			if errors.Is(err, notification.ErrSubscriptionExpired) {
				if unregisterErr := w.pushTokenService.UnregisterToken(ctx, pushToken.Token); unregisterErr != nil {
					slog.Warn("unregistering expired push subscription",
						slog.String("user_id", userID.String()),
						slog.Any("error", unregisterErr))
				}
			}
			continue
		}
		w.recordNotificationMetrics(ctx, "success")
		anySuccess = true
	}

	if anySuccess {
		slog.Info("push notification sent",
			slog.String("user_id", userID.String()),
			slog.String("notification", w.config.Name))
	}
}

func (w *PushNotificationWorker) sendToTenantUsers(ctx context.Context, tenantID domain.ID, msg async.BrokerMessage) {
	users, err := w.userService.FindByTenant(ctx, tenantID)
	if err != nil {
		slog.Error("failed to find users for tenant",
			slog.String("tenant_id", tenantID.String()),
			slog.Any("error", err))
		return
	}

	for _, user := range users {
		w.sendToUser(ctx, user.ID, msg)
	}
}

func (w *PushNotificationWorker) buildTitle(msg async.BrokerMessage) string {
	if w.config.TitleTemplate != "" {
		interpolated := w.interpolateTemplate(w.config.TitleTemplate, msg.Value)
		if !strings.Contains(interpolated, "{{") {
			return interpolated
		}
	}
	return w.config.Title
}

func (w *PushNotificationWorker) buildBody(msg async.BrokerMessage) string {
	return w.buildFromTemplate(w.config.BodyTemplate, w.config.Body, msg)
}

func (w *PushNotificationWorker) buildDeepLink(msg async.BrokerMessage) string {
	return w.buildFromTemplate(w.config.DeepLinkTemplate, w.config.DeepLink, msg)
}

// buildFromTemplate falls back to the static value when the payload did not
// supply every placeholder, so a missing key never reaches the user as a raw
// "{{key}}" in the notification body or as a broken deep link.
func (w *PushNotificationWorker) buildFromTemplate(template, fallback string, msg async.BrokerMessage) string {
	if template == "" {
		return fallback
	}

	interpolated := w.interpolateTemplate(template, msg.Value)
	if strings.Contains(interpolated, "{{") {
		slog.Warn("push notification template left a placeholder unresolved, falling back",
			slog.String("notification", w.config.Name),
			slog.String("result", interpolated))
		return fallback
	}

	return interpolated
}

// interpolateTemplate replaces every {{key}} placeholder with the matching value
// from the message payload. A placeholder with no matching key is left in place,
// which is how the caller detects an incomplete interpolation.
func (w *PushNotificationWorker) interpolateTemplate(template string, data any) string {
	result := template

	mapData, ok := data.(map[string]any)
	if !ok {
		return result
	}

	for key, value := range mapData {
		placeholder := "{{" + key + "}}"
		if strings.Contains(result, placeholder) {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}

	return result
}

func (w *PushNotificationWorker) initializeMetrics() error {
	meter := otel.Meter("push-notification-worker")

	notificationCounter, err := meter.Float64Counter(
		"zensor_server_push_notifications_total",
		metric.WithDescription("Total number of push notifications sent"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("creating notification counter: %w", err)
	}

	w.metricCounters[_metricKeyPushNotifications] = notificationCounter
	return nil
}

func (w *PushNotificationWorker) recordNotificationMetrics(ctx context.Context, status string) {
	if counter, exists := w.metricCounters[_metricKeyPushNotifications]; exists {
		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("notification", w.config.Name),
			attribute.String("status", status),
			semconv.ServiceNameKey.String("zensor-server"),
			semconv.ServiceVersionKey.String("1.0.0"),
		))
	}
}
