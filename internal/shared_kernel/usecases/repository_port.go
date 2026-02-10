package usecases

import (
	"context"
	"errors"
	"zensor-server/internal/shared_kernel/domain"
)

//go:generate mockgen -source=repository_port.go -destination=../../../test/unit/doubles/shared_kernel/usecases/repository_port_mock.go -package=usecases -mock_names=UserRepository=MockUserRepository,TenantRepository=MockTenantRepository,TenantConfigurationRepository=MockTenantConfigurationRepository,PushTokenRepository=MockPushTokenRepository

var (
	ErrTenantConfigurationNotFound = errors.New("tenant configuration not found")
	ErrUserNotFound                = errors.New("user not found")
	ErrPushTokenNotFound           = errors.New("push token not found")
	ErrTenantNotFound              = errors.New("tenant not found")
	ErrTenantDuplicated            = errors.New("tenant already exists")
	ErrTenantSoftDeleted           = errors.New("tenant is soft deleted")
	ErrTenantVersionConflict       = errors.New("tenant version conflict")
	ErrMixedTenantValidation       = errors.New("mixed tenant validation failed")
)

// Pagination encapsulates pagination parameters for repository queries
type Pagination struct {
	Limit  int
	Offset int
}

type UserRepository interface {
	Upsert(context.Context, domain.User) error
	GetByID(context.Context, domain.ID) (domain.User, error)
	FindByTenant(ctx context.Context, tenantID domain.ID) ([]domain.User, error)
}

type TenantRepository interface {
	Create(ctx context.Context, tenant domain.Tenant) error
	GetByID(ctx context.Context, id domain.ID) (domain.Tenant, error)
	GetByName(ctx context.Context, name string) (domain.Tenant, error)
	Update(ctx context.Context, tenant domain.Tenant) error
	FindAll(ctx context.Context, includeDeleted bool, pagination Pagination) ([]domain.Tenant, int, error)
}

type TenantConfigurationRepository interface {
	Create(context.Context, domain.TenantConfiguration) error
	GetByTenantID(context.Context, domain.ID) (domain.TenantConfiguration, error)
	Update(context.Context, domain.TenantConfiguration) error
}

type PushTokenRepository interface {
	Upsert(ctx context.Context, pushToken domain.PushToken) error
	GetByUserID(ctx context.Context, userID domain.ID) (domain.PushToken, error)
	DeleteByToken(ctx context.Context, token string) error
}
