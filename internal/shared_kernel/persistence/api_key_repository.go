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

func NewAPIKeyRepository(orm sql.ORM) (*SimpleAPIKeyRepository, error) {
	err := orm.AutoMigrate(&internal.APIKey{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleAPIKeyRepository{
		orm: orm,
	}, nil
}

var _ usecases.APIKeyRepository = (*SimpleAPIKeyRepository)(nil)

type SimpleAPIKeyRepository struct {
	orm sql.ORM
}

func (r *SimpleAPIKeyRepository) Create(ctx context.Context, key domain.APIKey) error {
	entity := internal.FromAPIKey(key)
	err := r.orm.WithContext(ctx).Create(&entity).Error()
	if err != nil {
		return fmt.Errorf("creating api key in database: %w", err)
	}

	return nil
}

func (r *SimpleAPIKeyRepository) GetByHash(ctx context.Context, keyHash string) (domain.APIKey, error) {
	return r.getByField(ctx, "key_hash = ?", keyHash)
}

func (r *SimpleAPIKeyRepository) GetByID(ctx context.Context, id domain.ID) (domain.APIKey, error) {
	return r.getByField(ctx, "id = ?", id.String())
}

func (r *SimpleAPIKeyRepository) GetByName(ctx context.Context, name string) (domain.APIKey, error) {
	return r.getByField(ctx, "name = ?", name)
}

func (r *SimpleAPIKeyRepository) getByField(ctx context.Context, query string, value string) (domain.APIKey, error) {
	var entity internal.APIKey
	err := r.orm.
		WithContext(ctx).
		Where(query, value).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.APIKey{}, usecases.ErrAPIKeyNotFound
	}

	if err != nil {
		return domain.APIKey{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleAPIKeyRepository) FindAll(ctx context.Context) ([]domain.APIKey, error) {
	var entities []internal.APIKey
	err := r.orm.
		WithContext(ctx).
		Order("name").
		Find(&entities).
		Error()
	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	keys := make([]domain.APIKey, 0, len(entities))
	for _, entity := range entities {
		keys = append(keys, entity.ToDomain())
	}

	return keys, nil
}

func (r *SimpleAPIKeyRepository) Delete(ctx context.Context, id domain.ID) error {
	err := r.orm.
		WithContext(ctx).
		Where("id = ?", id.String()).
		Delete(&internal.APIKey{}).
		Error()
	if err != nil {
		return fmt.Errorf("deleting api key from database: %w", err)
	}

	return nil
}
