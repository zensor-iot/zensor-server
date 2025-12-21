package notification

import (
	"context"
)

type CompositeNotificationClient struct {
	emailClient NotificationClient
	pushClient  NotificationClient
}

func NewCompositeNotificationClient(emailClient NotificationClient, pushClient NotificationClient) *CompositeNotificationClient {
	return &CompositeNotificationClient{
		emailClient: emailClient,
		pushClient:  pushClient,
	}
}

func (c *CompositeNotificationClient) SendEmail(ctx context.Context, request EmailRequest) error {
	return c.emailClient.SendEmail(ctx, request)
}

func (c *CompositeNotificationClient) SendPushNotification(ctx context.Context, request PushNotificationRequest) error {
	return c.pushClient.SendPushNotification(ctx, request)
}

