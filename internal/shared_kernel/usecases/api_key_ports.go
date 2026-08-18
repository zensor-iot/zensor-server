package usecases

import (
	"context"
	"errors"

	"zensor-server/internal/shared_kernel/domain"
)

//go:generate mockgen -source=api_key_ports.go -destination=../../../test/unit/doubles/shared_kernel/usecases/api_key_ports_mock.go -package=usecases -mock_names=APIKeyRepository=MockAPIKeyRepository,APIKeyService=MockAPIKeyService

var (
	ErrAPIKeyNotFound   = errors.New("api key not found")
	ErrAPIKeyDuplicated = errors.New("api key already exists")
)

type APIKeyRepository interface {
	Create(context.Context, domain.APIKey) error
	GetByHash(context.Context, string) (domain.APIKey, error)
	GetByID(context.Context, domain.ID) (domain.APIKey, error)
	GetByName(context.Context, string) (domain.APIKey, error)
	FindAll(context.Context) ([]domain.APIKey, error)
	Delete(context.Context, domain.ID) error
}

// APIKeyService manages machine credentials: creation, bearer validation,
// listing, and revocation.
type APIKeyService interface {
	Create(ctx context.Context, name string, createdBy domain.ID) (domain.APIKey, string, error)
	Validate(ctx context.Context, rawKey string) (domain.APIKey, error)
	List(ctx context.Context) ([]domain.APIKey, error)
	Revoke(ctx context.Context, id domain.ID) error
}
