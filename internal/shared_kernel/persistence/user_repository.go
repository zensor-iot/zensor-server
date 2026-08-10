package persistence

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"zensor-server/internal/shared_kernel/persistence/internal"
	"zensor-server/internal/shared_kernel/usecases"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/domain"
)

func NewUserRepository(orm sql.ORM) (*SimpleUserRepository, error) {
	err := orm.AutoMigrate(&internal.User{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleUserRepository{
		orm: orm,
	}, nil
}

var _ usecases.UserRepository = (*SimpleUserRepository)(nil)

type SimpleUserRepository struct {
	orm sql.ORM
}

func (r *SimpleUserRepository) Upsert(ctx context.Context, user domain.User) error {
	entity := internal.FromUser(user)

	var existing internal.User
	err := r.orm.WithContext(ctx).
		First(&existing, "id = ?", entity.ID).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		err = r.orm.WithContext(ctx).Create(&entity).Error()
		if err != nil {
			return fmt.Errorf("creating user in database: %w", err)
		}
		slog.Info("created user", slog.String("user_id", user.ID.String()), slog.Any("tenants", entity.Tenants))
		return nil
	}

	if err != nil {
		return fmt.Errorf("checking existing user: %w", err)
	}

	existing.Tenants = entity.Tenants
	existing.UpdatedAt = entity.UpdatedAt
	err = r.orm.WithContext(ctx).Save(&existing).Error()
	if err != nil {
		return fmt.Errorf("updating user in database: %w", err)
	}
	slog.Info("updated user", slog.String("user_id", user.ID.String()), slog.Any("tenants", entity.Tenants))

	return nil
}

func (r *SimpleUserRepository) GetByID(ctx context.Context, id domain.ID) (domain.User, error) {
	slog.Info("getting user by ID", slog.String("user_id", id.String()))
	var entity internal.User
	err := r.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id.String()).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		slog.Warn("user not found in database", slog.String("user_id", id.String()))
		return domain.User{}, usecases.ErrUserNotFound
	}

	if err != nil {
		slog.Error("database query error", slog.String("error", err.Error()))
		return domain.User{}, fmt.Errorf("database query: %w", err)
	}

	slog.Info("found user in database", slog.String("user_id", id.String()), slog.Any("tenants", entity.Tenants))
	return entity.ToDomain(), nil
}

func (r *SimpleUserRepository) FindByTenant(ctx context.Context, tenantID domain.ID) ([]domain.User, error) {
	var entities []internal.User
	err := r.orm.
		WithContext(ctx).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	users := make([]domain.User, 0, len(entities))
	for _, entity := range entities {
		user := entity.ToDomain()
		if slices.Contains(user.Tenants, tenantID) {
			users = append(users, user)
		}
	}

	return users, nil
}
