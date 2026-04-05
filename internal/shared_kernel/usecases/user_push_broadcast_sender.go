package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"zensor-server/internal/infra/notification"
	"zensor-server/internal/shared_kernel/domain"
)

// ErrUserPushBroadcastBodyRequired is returned when the broadcast body is empty or whitespace-only.
var ErrUserPushBroadcastBodyRequired = errors.New("push broadcast body is required")

// NewUserPushMessageSender builds a sender that resolves tokens via PushTokenService and delivers via notifier.
func NewUserPushMessageSender(pushTokens PushTokenService, notifier notification.NotificationClient) *SimpleUserPushMessageSender {
	return &SimpleUserPushMessageSender{
		pushTokens: pushTokens,
		notifier:   notifier,
	}
}

var _ UserPushMessageSender = (*SimpleUserPushMessageSender)(nil)

// SimpleUserPushMessageSender broadcasts FCM payloads to every token returned for the user.
type SimpleUserPushMessageSender struct {
	pushTokens PushTokenService
	notifier   notification.NotificationClient
}

// SendBroadcastToUser validates body, loads all tokens for the user, and sends best-effort per device.
func (s *SimpleUserPushMessageSender) SendBroadcastToUser(ctx context.Context, userID domain.ID, content UserPushBroadcastContent) error {
	body := strings.TrimSpace(content.Body)
	if body == "" {
		return ErrUserPushBroadcastBodyRequired
	}

	tokens, err := s.pushTokens.ListTokensByUserID(ctx, userID)
	if err != nil {
		return err
	}

	title := content.Title
	deepLink := content.DeepLink

	var successCount int
	var lastSendErr error
	for _, pt := range tokens {
		req := notification.PushNotificationRequest{
			Token:    pt.Token,
			Title:    title,
			Body:     body,
			DeepLink: deepLink,
		}
		sendErr := s.notifier.SendPushNotification(ctx, req)
		if sendErr != nil {
			lastSendErr = sendErr
			slog.Warn("push broadcast send failed",
				slog.String("user_id", userID.String()),
				slog.String("token_id", pt.ID.String()),
				slog.String("platform", pt.Platform),
				slog.String("error", sendErr.Error()))
			continue
		}
		successCount++
	}

	if successCount == 0 {
		if lastSendErr != nil {
			return fmt.Errorf("all push notification sends failed: %w", lastSendErr)
		}
		return fmt.Errorf("all push notification sends failed")
	}

	return nil
}
