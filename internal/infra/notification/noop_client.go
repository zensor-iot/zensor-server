package notification

import "context"

// NoOpClient is a NotificationClient that performs no I/O. It is used when
// ENV=local so the API process can run without FCM service account credentials.
type NoOpClient struct{}

// NewNoOpClient returns a NotificationClient that performs no network I/O.
func NewNoOpClient() *NoOpClient {
	return &NoOpClient{}
}

// SendEmail implements NotificationClient.
func (c *NoOpClient) SendEmail(ctx context.Context, request EmailRequest) error {
	return nil
}

// SendPushNotification implements NotificationClient.
func (c *NoOpClient) SendPushNotification(ctx context.Context, request PushNotificationRequest) error {
	return nil
}

var _ NotificationClient = (*NoOpClient)(nil)
