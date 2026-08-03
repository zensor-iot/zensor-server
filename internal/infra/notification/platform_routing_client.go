package notification

import (
	"context"
	"fmt"
)

const _platformWeb = "web"

// PlatformRoutingPushClient routes push notifications by token platform: web subscriptions to the web push client, everything else to FCM.
type PlatformRoutingPushClient struct {
	fcm     NotificationClient
	webPush NotificationClient
}

// NewPlatformRoutingPushClient creates a push client that dispatches by PushNotificationRequest.Platform.
func NewPlatformRoutingPushClient(fcm NotificationClient, webPush NotificationClient) *PlatformRoutingPushClient {
	return &PlatformRoutingPushClient{
		fcm:     fcm,
		webPush: webPush,
	}
}

var _ NotificationClient = (*PlatformRoutingPushClient)(nil)

// SendPushNotification dispatches the request to the client matching its platform.
func (c *PlatformRoutingPushClient) SendPushNotification(ctx context.Context, request PushNotificationRequest) error {
	if request.Platform == _platformWeb {
		return c.webPush.SendPushNotification(ctx, request)
	}
	return c.fcm.SendPushNotification(ctx, request)
}

// SendEmail is not supported by the push routing client
func (c *PlatformRoutingPushClient) SendEmail(ctx context.Context, request EmailRequest) error {
	return &NotificationError{
		Message: "email notifications are not supported by the platform routing push client",
		Err:     fmt.Errorf("use MailerSend client for email notifications"),
	}
}
