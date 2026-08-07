package notification

import (
	"context"
	"log/slog"
)

// NewFCMPushClient builds an FCM push client from the given configuration. When
// credentials cannot be loaded the returned client reports the failure on every
// send, so unavailable FCM credentials degrade push delivery instead of
// preventing the server from starting.
func NewFCMPushClient(ctx context.Context, config FCMConfig) NotificationClient {
	client, err := NewFCMClient(ctx, config)
	if err != nil {
		slog.Error("fcm push notifications are disabled: credentials could not be loaded",
			slog.String("error", err.Error()),
			slog.String("service_account_path", config.ServiceAccountPath),
		)
		return &unavailableFCMClient{err: err}
	}

	return client
}

type unavailableFCMClient struct {
	err error
}

var _ NotificationClient = (*unavailableFCMClient)(nil)

func (c *unavailableFCMClient) SendPushNotification(ctx context.Context, request PushNotificationRequest) error {
	return &NotificationError{
		Message: "fcm push notifications are unavailable",
		Err:     c.err,
	}
}

func (c *unavailableFCMClient) SendEmail(ctx context.Context, request EmailRequest) error {
	return &NotificationError{
		Message: "email notifications are not supported by FCM",
		Err:     c.err,
	}
}
