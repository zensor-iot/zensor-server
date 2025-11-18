package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/mailersend/mailersend-go"
)

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

func (c *MailerSendConfig) validateConfig() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.FromEmail == "" {
		return fmt.Errorf("From email is required")
	}
	if c.FromName == "" {
		return fmt.Errorf("From name is required")
	}
	return nil
}

// NewMailerSendClient creates a new MailerSend client
func NewMailerSendClient(config MailerSendConfig) *MailerSendClient {
	if err := config.validateConfig(); err != nil {
		panic(err)
	}

	client := mailersend.NewMailersend(config.APIKey)

	return &MailerSendClient{
		client:    client,
		fromEmail: config.FromEmail,
		fromName:  config.FromName,
	}
}

// SendEmail sends an email using MailerSend API
func (c *MailerSendClient) SendEmail(ctx context.Context, request EmailRequest) error {
	message := c.client.Email.NewMessage()

	message.SetFrom(mailersend.From{
		Email: c.fromEmail,
		Name:  c.fromName,
	})

	message.SetRecipients([]mailersend.Recipient{
		{
			Email: request.To,
		},
	})

	message.SetSubject(request.Subject)
	message.SetText(request.Body)

	return c.sendWithRetry(ctx, message)
}

func (c *MailerSendClient) sendWithRetry(ctx context.Context, message *mailersend.Message) error {
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		_, err := c.client.Email.Send(ctx, message)
		if err != nil {
			lastErr = &NotificationError{
				Message: fmt.Sprintf("MailerSend API error (attempt %d/3)", attempt),
				Err:     err,
			}

			if attempt < 3 {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return lastErr
		}

		return nil
	}

	return lastErr
}
