package persistence

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/domain"
)

func NewPushTokenRepository(orm sql.ORM) (*SimplePushTokenRepository, error) {
	err := orm.AutoMigrate(&internal.PushToken{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimplePushTokenRepository{
		orm: orm,
	}, nil
}

var _ usecases.PushTokenRepository = (*SimplePushTokenRepository)(nil)

type SimplePushTokenRepository struct {
	orm sql.ORM
}

func (r *SimplePushTokenRepository) Upsert(ctx context.Context, pushToken domain.PushToken) error {
	entity := internal.FromPushToken(pushToken)
	entity.UpdatedAt = pushToken.UpdatedAt.Time

	var existing internal.PushToken
	err := r.orm.WithContext(ctx).
		Where("user_id = ? AND token = ?", entity.UserID, entity.Token).
		First(&existing).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		err = r.orm.WithContext(ctx).Create(&entity).Error()
		if err != nil {
			return fmt.Errorf("creating push token: %w", err)
		}
		slog.Info("created push token",
			slog.String("user_id", pushToken.UserID.String()),
			slog.String("token_id", pushToken.ID.String()),
			slog.String("platform", pushToken.Platform))
		return nil
	}

	if err != nil {
		return fmt.Errorf("checking existing push token: %w", err)
	}

	existing.Platform = entity.Platform
	existing.UpdatedAt = entity.UpdatedAt
	err = r.orm.WithContext(ctx).Save(&existing).Error()
	if err != nil {
		return fmt.Errorf("updating push token: %w", err)
	}

	slog.Info("updated push token",
		slog.String("user_id", pushToken.UserID.String()),
		slog.String("token_id", pushToken.ID.String()),
		slog.String("platform", pushToken.Platform))

	return nil
}

func (r *SimplePushTokenRepository) GetByUserID(ctx context.Context, userID domain.ID) (domain.PushToken, error) {
	var entity internal.PushToken
	err := r.orm.
		WithContext(ctx).
		Where("user_id = ?", userID.String()).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.PushToken{}, usecases.ErrPushTokenNotFound
	}

	if err != nil {
		return domain.PushToken{}, fmt.Errorf("getting push token: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimplePushTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	var existing internal.PushToken
	err := r.orm.WithContext(ctx).
		Where("token = ?", token).
		First(&existing).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return usecases.ErrPushTokenNotFound
	}

	if err != nil {
		return fmt.Errorf("finding push token: %w", err)
	}

	err = r.orm.WithContext(ctx).
		Delete(&existing).
		Error()

	if err != nil {
		return fmt.Errorf("deleting push token: %w", err)
	}

	slog.Info("deleted push token", slog.String("token", token))
	return nil
}

