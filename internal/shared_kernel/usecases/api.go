package usecases

import (
	"context"
	"zensor-server/internal/shared_kernel/domain"
)

//go:generate mockgen -source=./api.go -destination=../../../test/unit/doubles/shared_kernel/usecases/api_mock.go -package=usecases

type UserService interface {
	AssociateTenants(context.Context, domain.ID, []domain.ID) error
	GetUser(context.Context, domain.ID) (domain.User, error)
	FindByTenant(context.Context, domain.ID) ([]domain.User, error)
}

type TenantConfigurationService interface {
	UpsertTenantConfiguration(ctx context.Context, userEmail string, config domain.TenantConfiguration) (domain.TenantConfiguration, error)
	GetTenantConfiguration(ctx context.Context, tenant domain.Tenant) (domain.TenantConfiguration, error)
	GetOrCreateTenantConfiguration(ctx context.Context, tenant domain.Tenant, defaultTimezone string) (domain.TenantConfiguration, error)
}

type TenantService interface {
	CreateTenant(ctx context.Context, tenant domain.Tenant) error
	GetTenant(ctx context.Context, id domain.ID) (domain.Tenant, error)
	GetTenantByName(ctx context.Context, name string) (domain.Tenant, error)
	ListTenants(ctx context.Context, includeDeleted bool, pagination Pagination) ([]domain.Tenant, int, error)
	UpdateTenant(ctx context.Context, tenant domain.Tenant) error
	SoftDeleteTenant(ctx context.Context, id domain.ID) error
	ActivateTenant(ctx context.Context, id domain.ID) error
	DeactivateTenant(ctx context.Context, id domain.ID) error
	AdoptDevice(ctx context.Context, tenantID, deviceID domain.ID) error
	ListTenantDevices(ctx context.Context, tenantID domain.ID, pagination Pagination) ([]domain.Device, int, error)
}

type PushTokenService interface {
	RegisterToken(ctx context.Context, userID domain.ID, token string, platform string) error
	UnregisterToken(ctx context.Context, token string) error
	GetTokenByUserID(ctx context.Context, userID domain.ID) (domain.PushToken, error)
}

type DeviceAdopter interface {
	AdoptDeviceToTenant(ctx context.Context, tenantID, deviceID domain.ID) error
	DevicesByTenant(ctx context.Context, tenantID domain.ID, pagination Pagination) ([]domain.Device, int, error)
}
