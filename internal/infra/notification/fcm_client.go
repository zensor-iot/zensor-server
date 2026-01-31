package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	_ "cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	_fcmEndpoint        = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
	_fcmScope           = "https://www.googleapis.com/auth/firebase.messaging"
	_maxRetries         = 3
	_tokenRefreshBuffer = 5 * time.Minute
)

type FCMClient struct {
	httpClient  *http.Client
	projectID   string
	tokenSource oauth2.TokenSource
	mu          sync.Mutex
	cachedToken *oauth2.Token
}

type FCMConfig struct {
	ProjectID          string
	ServiceAccountPath string
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

func NewFCMClient(ctx context.Context, config FCMConfig) (*FCMClient, error) {
	var creds *google.Credentials
	var err error
	if config.ServiceAccountPath != "" {
		jsonBytes, err := os.ReadFile(config.ServiceAccountPath)
		if err != nil {
			return nil, fmt.Errorf("reading service account file: %w", err)
		}
		creds, err = google.CredentialsFromJSON(ctx, jsonBytes, _fcmScope)
		if err != nil {
			return nil, fmt.Errorf("creating credentials from service account: %w", err)
		}
	} else {
		creds, err = google.FindDefaultCredentials(ctx, _fcmScope)
		if err != nil {
			return nil, fmt.Errorf("finding default credentials (set GOOGLE_APPLICATION_CREDENTIALS or fcm.service_account_path): %w", err)
		}
	}
	return &FCMClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		projectID:   config.ProjectID,
		tokenSource: creds.TokenSource,
	}, nil
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

func isTokenValid(token *oauth2.Token) bool {
	if token == nil {
		return false
	}
	if token.Expiry.IsZero() {
		return true
	}
	return time.Now().Add(_tokenRefreshBuffer).Before(token.Expiry)
}

func (c *FCMClient) getValidToken() (*oauth2.Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if isTokenValid(c.cachedToken) {
		return c.cachedToken, nil
	}
	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, err
	}
	c.cachedToken = token
	return token, nil
}

func (c *FCMClient) send(ctx context.Context, request FCMRequest) error {
	token, err := c.getValidToken()
	if err != nil {
		return fmt.Errorf("getting access token: %w", err)
	}

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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

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
