package notification

import "errors"

// Sentinel errors returned by notification clients.
var (
	ErrEmailNotSupported   = errors.New("use MailerSend client for email notifications")
	ErrPushNotSupported    = errors.New("use FCM client for push notifications")
	ErrAPIKeyRequired      = errors.New("API key is required")
	ErrFromEmailRequired   = errors.New("from email is required")
	ErrFromNameRequired    = errors.New("from name is required")
	ErrFCMAPIError         = errors.New("FCM API error")
	ErrWebPushServiceError = errors.New("web push service error")
)
