package notification

import (
	"context"
)

//go:generate mockgen -source=notification_client.go -destination=../../../test/unit/doubles/infra/notification/notification_client_mock.go -package=notification -mock_names=NotificationClient=MockNotificationClient

// NotificationClient defines the interface for sending notifications
type NotificationClient interface {
	// SendEmail sends an email notification
	SendEmail(ctx context.Context, request EmailRequest) error
	// SendPushNotification sends a push notification
	SendPushNotification(ctx context.Context, request PushNotificationRequest) error
}

// EmailRequest represents the data needed to send an email notification
type EmailRequest struct {
	To      string
	Subject string
	Body    string
}

// PushNotificationRequest represents the data needed to send a push notification
type PushNotificationRequest struct {
	Token    string
	Title    string
	Body     string
	DeepLink string
}

// NotificationError represents an error that occurred during notification sending
type NotificationError struct {
	Message string
	Err     error
}

func (e *NotificationError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *NotificationError) Unwrap() error {
	return e.Err
}
