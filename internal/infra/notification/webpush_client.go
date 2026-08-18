package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// ErrSubscriptionExpired indicates the push service reported the subscription is gone and its token should be unregistered.
var ErrSubscriptionExpired = errors.New("push subscription expired")

// WebPushConfig holds VAPID credentials for the Web Push protocol.
type WebPushConfig struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	Subscriber      string
}

// WebPushClient implements NotificationClient using native Web Push with VAPID.
type WebPushClient struct {
	config     WebPushConfig
	httpClient *http.Client
}

// NewWebPushClient creates a Web Push client from VAPID configuration.
func NewWebPushClient(config WebPushConfig) *WebPushClient {
	return &WebPushClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

var _ NotificationClient = (*WebPushClient)(nil)

type webPushPayload struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	DeepLink string `json:"deeplink"`
}

// SendPushNotification delivers the notification to the browser subscription stored as JSON in request.Token.
func (c *WebPushClient) SendPushNotification(ctx context.Context, request PushNotificationRequest) error {
	var subscription webpush.Subscription
	if err := json.Unmarshal([]byte(request.Token), &subscription); err != nil {
		return &NotificationError{
			Message: "decoding web push subscription",
			Err:     err,
		}
	}

	payload, err := json.Marshal(webPushPayload{
		Title:    request.Title,
		Body:     request.Body,
		DeepLink: request.DeepLink,
	})
	if err != nil {
		return &NotificationError{
			Message: "encoding web push payload",
			Err:     err,
		}
	}

	resp, err := webpush.SendNotificationWithContext(ctx, payload, &subscription, &webpush.Options{
		HTTPClient:      c.httpClient,
		Subscriber:      c.config.Subscriber,
		VAPIDPublicKey:  c.config.VAPIDPublicKey,
		VAPIDPrivateKey: c.config.VAPIDPrivateKey,
		TTL:             3600,
	})
	if err != nil {
		return &NotificationError{
			Message: "sending web push notification",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return ErrSubscriptionExpired
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		return &NotificationError{
			Message: "web push service error",
			Err:     fmt.Errorf("unexpected status %d", resp.StatusCode),
		}
	}

	return nil
}

// SendEmail is not supported by Web Push.
func (c *WebPushClient) SendEmail(ctx context.Context, request EmailRequest) error {
	return &NotificationError{
		Message: "email notifications are not supported by Web Push",
		Err:     errors.New("use MailerSend client for email notifications"),
	}
}
