package notification

import (
	"context"
)

// NotificationClient defines the interface for sending notifications
type NotificationClient interface {
	// SendEmail sends an email notification
	SendEmail(ctx context.Context, request EmailRequest) error
}

// EmailRequest represents the data needed to send an email notification
type EmailRequest struct {
	To      string
	Subject string
	Body    string
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
