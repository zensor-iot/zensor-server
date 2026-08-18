package usecases

import (
	"context"

	"zensor-server/internal/shared_kernel/domain"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"
)

//go:generate mockgen -source=./api.go -destination=../../../test/unit/doubles/control_plane/usecases/api.go -package=usecases

type DeviceService interface {
	CreateDevice(context.Context, domain.Device) error
	GetDevice(context.Context, domain.ID) (domain.Device, error)
	AllDevices(context.Context, Pagination) ([]domain.Device, int, error)
	DevicesByTenant(context.Context, domain.ID, Pagination) ([]domain.Device, int, error)
	UpdateDeviceDisplayName(context.Context, domain.ID, string) error
	QueueCommand(context.Context, domain.Command) error
	QueueCommandSequence(context.Context, domain.CommandSequence) error
	AdoptDeviceToTenant(context.Context, domain.ID, domain.ID) error
	UpdateLastMessageReceivedAt(context.Context, string) error
}

type EvaluationRuleService interface {
	AddToDevice(context.Context, domain.Device, domain.EvaluationRule) error
	FindAllByDevice(context.Context, domain.Device) ([]domain.EvaluationRule, error)
}

type TaskService interface {
	Create(context.Context, domain.Task) error
	FindAllByDevice(context.Context, domain.ID, Pagination) ([]domain.Task, int, error)
	FindAllByScheduledTask(context.Context, domain.ID, Pagination) ([]domain.Task, int, error)
}

type ScheduledTaskService interface {
	Create(context.Context, domain.ScheduledTask) error
	FindAllByTenant(context.Context, domain.ID) ([]domain.ScheduledTask, error)
	FindAllByTenantAndDevice(context.Context, domain.ID, domain.ID, Pagination) ([]domain.ScheduledTask, int, error)
	GetByID(context.Context, domain.ID) (domain.ScheduledTask, error)
	Update(context.Context, domain.ScheduledTask) error
	Delete(context.Context, domain.ID) error
}

// Type aliases for types moved to shared_kernel/usecases.
type (
	UserService                      = sharedUsecases.UserService
	TenantService                    = sharedUsecases.TenantService
	TenantConfigurationService       = sharedUsecases.TenantConfigurationService
	PushTokenService                 = sharedUsecases.PushTokenService
	UserRepository                   = sharedUsecases.UserRepository
	TenantRepository                 = sharedUsecases.TenantRepository
	TenantConfigurationRepository    = sharedUsecases.TenantConfigurationRepository
	PushTokenRepository              = sharedUsecases.PushTokenRepository
	SimpleUserService                = sharedUsecases.SimpleUserService
	SimpleTenantService              = sharedUsecases.SimpleTenantService
	SimpleTenantConfigurationService = sharedUsecases.SimpleTenantConfigurationService
	SimplePushTokenService           = sharedUsecases.SimplePushTokenService
)

// Constructor aliases.
var (
	NewUserService                = sharedUsecases.NewUserService
	NewTenantService              = sharedUsecases.NewTenantService
	NewTenantConfigurationService = sharedUsecases.NewTenantConfigurationService
	NewPushTokenService           = sharedUsecases.NewPushTokenService
)

// Error aliases.
var (
	ErrTenantConfigurationNotFound        = sharedUsecases.ErrTenantConfigurationNotFound
	ErrUserNotFound                       = sharedUsecases.ErrUserNotFound
	ErrPushTokenNotFound                  = sharedUsecases.ErrPushTokenNotFound
	ErrTenantNotFound                     = sharedUsecases.ErrTenantNotFound
	ErrTenantDuplicated                   = sharedUsecases.ErrTenantDuplicated
	ErrTenantSoftDeleted                  = sharedUsecases.ErrTenantSoftDeleted
	ErrTenantVersionConflict              = sharedUsecases.ErrTenantVersionConflict
	ErrMixedTenantValidation              = sharedUsecases.ErrMixedTenantValidation
	ErrInvalidTimezone                    = sharedUsecases.ErrInvalidTimezone
	ErrForbiddenTenantConfigurationAccess = sharedUsecases.ErrForbiddenTenantConfigurationAccess
)
