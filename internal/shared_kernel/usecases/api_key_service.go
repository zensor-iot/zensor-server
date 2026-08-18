package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"
	"zensor-server/internal/infra/cache"
	"zensor-server/internal/shared_kernel/domain"
)

const (
	apiKeyCacheTTL       = 24 * time.Hour
	apiKeyCacheKeyPrefix = "apikey:"
)

func NewAPIKeyService(repository APIKeyRepository, keyCache cache.Cache) *SimpleAPIKeyService {
	return &SimpleAPIKeyService{
		repository: repository,
		cache:      keyCache,
	}
}

var _ APIKeyService = (*SimpleAPIKeyService)(nil)

type SimpleAPIKeyService struct {
	repository APIKeyRepository
	cache      cache.Cache
}

func (s *SimpleAPIKeyService) Create(ctx context.Context, name string, createdBy domain.ID) (domain.APIKey, string, error) {
	key, plaintext, err := domain.NewAPIKeyBuilder().
		WithName(name).
		WithCreatedBy(createdBy).
		Build()
	if err != nil {
		return domain.APIKey{}, "", err
	}

	_, err = s.repository.GetByName(ctx, key.Name)
	if err == nil {
		return domain.APIKey{}, "", ErrAPIKeyDuplicated
	}
	if !errors.Is(err, ErrAPIKeyNotFound) {
		return domain.APIKey{}, "", fmt.Errorf("checking api key name: %w", err)
	}

	if err := s.repository.Create(ctx, key); err != nil {
		return domain.APIKey{}, "", fmt.Errorf("creating api key: %w", err)
	}

	return key, plaintext, nil
}

func (s *SimpleAPIKeyService) Validate(ctx context.Context, rawKey string) (domain.APIKey, error) {
	hash := domain.HashAPIKey(rawKey)
	value, err := s.cache.GetOrSet(ctx, apiKeyCacheKeyPrefix+hash, apiKeyCacheTTL, func() (any, error) {
		return s.repository.GetByHash(ctx, hash)
	})
	if err != nil {
		return domain.APIKey{}, err
	}

	key, ok := value.(domain.APIKey)
	if !ok {
		return domain.APIKey{}, ErrAPIKeyNotFound
	}

	return key, nil
}

func (s *SimpleAPIKeyService) List(ctx context.Context) ([]domain.APIKey, error) {
	keys, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing api keys: %w", err)
	}

	return keys, nil
}

func (s *SimpleAPIKeyService) Revoke(ctx context.Context, id domain.ID) error {
	key, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting api key: %w", err)
	}

	s.cache.Delete(ctx, apiKeyCacheKeyPrefix+key.KeyHash)

	return nil
}
