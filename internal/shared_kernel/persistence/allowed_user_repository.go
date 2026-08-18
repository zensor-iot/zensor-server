// Package persistence provides repository implementations for shared kernel resources.
package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/persistence/internal"
	"zensor-server/internal/shared_kernel/usecases"
)

func NewAllowedUserRepository(orm sql.ORM) (*SimpleAllowedUserRepository, error) {
	err := orm.AutoMigrate(&internal.AllowedUser{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleAllowedUserRepository{
		orm: orm,
	}, nil
}

var _ usecases.AllowedUserRepository = (*SimpleAllowedUserRepository)(nil)

type SimpleAllowedUserRepository struct {
	orm sql.ORM
}

func (r *SimpleAllowedUserRepository) Create(ctx context.Context, user domain.AllowedUser) error {
	entity := internal.FromAllowedUser(user)
	err := r.orm.WithContext(ctx).Create(&entity).Error()
	if err != nil {
		return fmt.Errorf("creating allowed user in database: %w", err)
	}

	return nil
}

func (r *SimpleAllowedUserRepository) GetByID(ctx context.Context, id domain.ID) (domain.AllowedUser, error) {
	var entity internal.AllowedUser
	err := r.orm.
		WithContext(ctx).
		Where("id = ?", id.String()).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.AllowedUser{}, usecases.ErrAllowedUserNotFound
	}

	if err != nil {
		return domain.AllowedUser{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleAllowedUserRepository) GetByEmail(ctx context.Context, email string) (domain.AllowedUser, error) {
	var entity internal.AllowedUser
	err := r.orm.
		WithContext(ctx).
		Where("email = ?", domain.NormalizeEmail(email)).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.AllowedUser{}, usecases.ErrAllowedUserNotFound
	}

	if err != nil {
		return domain.AllowedUser{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleAllowedUserRepository) FindAll(ctx context.Context) ([]domain.AllowedUser, error) {
	var entities []internal.AllowedUser
	err := r.orm.
		WithContext(ctx).
		Order("email").
		Find(&entities).
		Error()
	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	users := make([]domain.AllowedUser, 0, len(entities))
	for _, entity := range entities {
		users = append(users, entity.ToDomain())
	}

	return users, nil
}

func (r *SimpleAllowedUserRepository) Update(ctx context.Context, user domain.AllowedUser) error {
	entity := internal.FromAllowedUser(user)
	err := r.orm.WithContext(ctx).Save(&entity).Error()
	if err != nil {
		return fmt.Errorf("updating allowed user in database: %w", err)
	}

	return nil
}

func (r *SimpleAllowedUserRepository) Delete(ctx context.Context, id domain.ID) error {
	err := r.orm.
		WithContext(ctx).
		Where("id = ?", id.String()).
		Delete(&internal.AllowedUser{}).
		Error()
	if err != nil {
		return fmt.Errorf("deleting allowed user from database: %w", err)
	}

	return nil
}
