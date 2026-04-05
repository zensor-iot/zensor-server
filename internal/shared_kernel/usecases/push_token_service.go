package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
)

func NewPushTokenService(repository PushTokenRepository) *SimplePushTokenService {
	return &SimplePushTokenService{
		repository: repository,
	}
}

var _ PushTokenService = &SimplePushTokenService{}

type SimplePushTokenService struct {
	repository PushTokenRepository
}

func (s *SimplePushTokenService) RegisterToken(ctx context.Context, userID domain.ID, token string, platform string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	if token == "" {
		return errors.New("token is required")
	}
	if platform == "" {
		return errors.New("platform is required")
	}

	pushToken, err := domain.NewPushTokenBuilder().
		WithUserID(userID).
		WithToken(token).
		WithPlatform(platform).
		Build()
	if err != nil {
		return fmt.Errorf("building push token: %w", err)
	}

	now := utils.Time{Time: time.Now()}
	pushToken.UpdatedAt = now

	err = s.repository.Upsert(ctx, pushToken)
	if err != nil {
		slog.Error("registering push token", slog.String("error", err.Error()))
		return fmt.Errorf("registering push token: %w", err)
	}

	slog.Info("push token registered",
		slog.String("user_id", userID.String()),
		slog.String("token_id", pushToken.ID.String()),
		slog.String("platform", platform))

	return nil
}

func (s *SimplePushTokenService) UnregisterToken(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is required")
	}

	err := s.repository.DeleteByToken(ctx, token)
	if errors.Is(err, ErrPushTokenNotFound) {
		slog.Warn("push token not found for unregistration", slog.String("token", token))
		return ErrPushTokenNotFound
	}
	if err != nil {
		slog.Error("unregistering push token", slog.String("error", err.Error()))
		return fmt.Errorf("unregistering push token: %w", err)
	}

	slog.Info("push token unregistered", slog.String("token", token))
	return nil
}

func (s *SimplePushTokenService) GetTokenByUserID(ctx context.Context, userID domain.ID) (domain.PushToken, error) {
	if userID == "" {
		return domain.PushToken{}, errors.New("user ID is required")
	}

	token, err := s.repository.GetByUserID(ctx, userID)
	if errors.Is(err, ErrPushTokenNotFound) {
		return domain.PushToken{}, ErrPushTokenNotFound
	}
	if err != nil {
		slog.Error("getting push token", slog.String("error", err.Error()))
		return domain.PushToken{}, fmt.Errorf("getting push token: %w", err)
	}

	return token, nil
}

// ListTokensByUserID returns all push tokens for the user, or ErrPushTokenNotFound if the list is empty.
func (s *SimplePushTokenService) ListTokensByUserID(ctx context.Context, userID domain.ID) ([]domain.PushToken, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	tokens, err := s.repository.ListByUserID(ctx, userID)
	if err != nil {
		slog.Error("listing push tokens", slog.String("error", err.Error()))
		return nil, fmt.Errorf("listing push tokens: %w", err)
	}
	if len(tokens) == 0 {
		return nil, ErrPushTokenNotFound
	}
	return tokens, nil
}
