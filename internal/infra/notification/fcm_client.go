package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	_fcmEndpoint = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
	_maxRetries  = 3
)

type FCMClient struct {
	httpClient  *http.Client
	projectID   string
	accessToken string
}

type FCMConfig struct {
	ProjectID   string
	AccessToken string
}

type FCMRequest struct {
	Message FCMessage `json:"message"`
}

type FCMessage struct {
	Token        string            `json:"token"`
	Notification *FCNotification   `json:"notification,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
	Android      *FCAndroidConfig  `json:"android,omitempty"`
}

type FCNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type FCAndroidConfig struct {
	Priority string `json:"priority"`
}

func NewFCMClient(config FCMConfig) *FCMClient {
	return &FCMClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		projectID:   config.ProjectID,
		accessToken: config.AccessToken,
	}
}

func (c *FCMClient) SendPushNotification(ctx context.Context, request PushNotificationRequest) error {
	fcmRequest := FCMRequest{
		Message: FCMessage{
			Token: request.Token,
			Notification: &FCNotification{
				Title: request.Title,
				Body:  request.Body,
			},
			Android: &FCAndroidConfig{
				Priority: "high",
			},
		},
	}

	if request.DeepLink != "" {
		if fcmRequest.Message.Data == nil {
			fcmRequest.Message.Data = make(map[string]string)
		}
		fcmRequest.Message.Data["deeplink"] = request.DeepLink
	}

	return c.sendWithRetry(ctx, fcmRequest)
}

func (c *FCMClient) sendWithRetry(ctx context.Context, request FCMRequest) error {
	var lastErr error

	for attempt := 1; attempt <= _maxRetries; attempt++ {
		err := c.send(ctx, request)
		if err != nil {
			lastErr = &NotificationError{
				Message: fmt.Sprintf("FCM API error (attempt %d/%d)", attempt, _maxRetries),
				Err:     err,
			}

			if attempt < _maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return lastErr
		}

		return nil
	}

	return lastErr
}

func (c *FCMClient) send(ctx context.Context, request FCMRequest) error {
	url := fmt.Sprintf(_fcmEndpoint, c.projectID)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshaling FCM request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("FCM API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendEmail is not supported by FCM
func (c *FCMClient) SendEmail(ctx context.Context, request EmailRequest) error {
	return &NotificationError{
		Message: "email notifications are not supported by FCM",
		Err:     fmt.Errorf("use MailerSend client for email notifications"),
	}
}
