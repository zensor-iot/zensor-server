package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/mailersend/mailersend-go"
)

//go:generate mockgen -source=mailersend_client.go -destination=../../../test/unit/doubles/infra/notification/mailersend_client_mock.go -package=notification -mock_names=MailerSendClient=MockMailerSendClient

// MailerSendClient implements NotificationClient using MailerSend API
type MailerSendClient struct {
	client    *mailersend.Mailersend
	fromEmail string
	fromName  string
}

// MailerSendConfig holds configuration for MailerSend client
type MailerSendConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
}

// NewMailerSendClient creates a new MailerSend client
func NewMailerSendClient(config MailerSendConfig) *MailerSendClient {
	client := mailersend.NewMailersend(config.APIKey)

	return &MailerSendClient{
		client:    client,
		fromEmail: config.FromEmail,
		fromName:  config.FromName,
	}
}

// SendEmail sends an email using MailerSend API
func (c *MailerSendClient) SendEmail(ctx context.Context, request EmailRequest) error {
	// Create email message using the official library
	message := c.client.Email.NewMessage()

	// Set sender
	message.SetFrom(mailersend.From{
		Email: c.fromEmail,
		Name:  c.fromName,
	})

	// Set recipient
	message.SetRecipients([]mailersend.Recipient{
		{
			Email: request.To,
		},
	})

	// Set subject and text content
	message.SetSubject(request.Subject)
	message.SetText(request.Body)

	// Send email with retry logic
	return c.sendWithRetry(ctx, message)
}

// sendWithRetry sends the email with retry logic (3 attempts)
func (c *MailerSendClient) sendWithRetry(ctx context.Context, message *mailersend.Message) error {
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		// Send the email using the official library
		_, err := c.client.Email.Send(ctx, message)
		if err != nil {
			lastErr = &NotificationError{
				Message: fmt.Sprintf("MailerSend API error (attempt %d/3)", attempt),
				Err:     err,
			}

			if attempt < 3 {
				// Wait before retry (exponential backoff)
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return lastErr
		}

		// Success
		return nil
	}

	return lastErr
}

// SendPushNotification is not supported by MailerSend
func (c *MailerSendClient) SendPushNotification(ctx context.Context, request PushNotificationRequest) error {
	return &NotificationError{
		Message: "push notifications are not supported by MailerSend",
		Err:     fmt.Errorf("use FCM client for push notifications"),
	}
}
