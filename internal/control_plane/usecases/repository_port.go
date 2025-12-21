package usecases

import (
	"context"
	"errors"
	"zensor-server/internal/shared_kernel/domain"
)

//go:generate mockgen -source=repository_port.go -destination=../../../test/unit/doubles/control_plane/usecases/repository_port_mock.go -package=usecases -mock_names=DeviceRepository=MockDeviceRepository,CommandRepository=MockCommandRepository,EvaluationRuleRepository=MockEvaluationRuleRepository,TaskRepository=MockTaskRepository,ScheduledTaskRepository=MockScheduledTaskRepository,TenantConfigurationRepository=MockTenantConfigurationRepository,UserRepository=MockUserRepository

var (
	ErrDeviceNotFound              = errors.New("device not found")
	ErrDeviceDuplicated            = errors.New("device already exists")
	ErrCommandOverlap              = errors.New("command overlap detected")
	ErrTenantConfigurationNotFound = errors.New("tenant configuration not found")
	ErrUserNotFound                = errors.New("user not found")
	ErrPushTokenNotFound           = errors.New("push token not found")
)

// Pagination encapsulates pagination parameters for repository queries
type Pagination struct {
	Limit  int
	Offset int
}

type DeviceRepository interface {
	CreateDevice(context.Context, domain.Device) error
	UpdateDevice(context.Context, domain.Device) error
	Get(context.Context, string) (domain.Device, error)
	FindByName(context.Context, string) (domain.Device, error)
	FindAll(context.Context, Pagination) ([]domain.Device, int, error)
	FindByTenant(context.Context, string, Pagination) ([]domain.Device, int, error)
	AddEvaluationRule(context.Context, domain.Device, domain.EvaluationRule) error
	FindAllEvaluationRules(context.Context, domain.Device) ([]domain.EvaluationRule, error)
}

type CommandRepository interface {
	Create(context.Context, domain.Command) error
	Update(context.Context, domain.Command) error
	GetByID(context.Context, domain.ID) (domain.Command, error)
	FindAllPending(context.Context) ([]domain.Command, error)
	FindPendingByDevice(context.Context, domain.ID) ([]domain.Command, error)
	FindByTaskID(context.Context, domain.ID) ([]domain.Command, error)
}

type EvaluationRuleRepository interface {
	AddToDevice(context.Context, domain.Device, domain.EvaluationRule) error
	FindAllByDeviceID(ctx context.Context, deviceID string) ([]domain.EvaluationRule, error)
}

type TaskRepository interface {
	Create(context.Context, domain.Task) error
	FindAllByDevice(ctx context.Context, device domain.Device, pagination Pagination) ([]domain.Task, int, error)
	FindAllByScheduledTask(ctx context.Context, scheduledTaskID domain.ID, pagination Pagination) ([]domain.Task, int, error)
}

type ScheduledTaskRepository interface {
	Create(context.Context, domain.ScheduledTask) error
	FindAllByTenant(context.Context, domain.ID) ([]domain.ScheduledTask, error)
	FindAllByTenantAndDevice(context.Context, domain.ID, domain.ID, Pagination) ([]domain.ScheduledTask, int, error)
	FindAllActive(context.Context) ([]domain.ScheduledTask, error)
	Update(context.Context, domain.ScheduledTask) error
	GetByID(context.Context, domain.ID) (domain.ScheduledTask, error)
	Delete(context.Context, domain.ID) error
}

type TenantConfigurationRepository interface {
	Create(context.Context, domain.TenantConfiguration) error
	GetByTenantID(context.Context, domain.ID) (domain.TenantConfiguration, error)
	Update(context.Context, domain.TenantConfiguration) error
}

type UserRepository interface {
	Upsert(context.Context, domain.User) error
	GetByID(context.Context, domain.ID) (domain.User, error)
}

type TenantRepository interface {
	Create(ctx context.Context, tenant domain.Tenant) error
	GetByID(ctx context.Context, id domain.ID) (domain.Tenant, error)
	GetByName(ctx context.Context, name string) (domain.Tenant, error)
	Update(ctx context.Context, tenant domain.Tenant) error
	FindAll(ctx context.Context, includeDeleted bool, pagination Pagination) ([]domain.Tenant, int, error)
}

type PushTokenRepository interface {
	Upsert(ctx context.Context, pushToken domain.PushToken) error
	GetByUserID(ctx context.Context, userID domain.ID) (domain.PushToken, error)
	DeleteByToken(ctx context.Context, token string) error
}
