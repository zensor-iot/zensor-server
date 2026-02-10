package usecases

import (
	"context"
	"errors"
	"zensor-server/internal/shared_kernel/domain"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"
)

//go:generate mockgen -source=repository_port.go -destination=../../../test/unit/doubles/control_plane/usecases/repository_port_mock.go -package=usecases -mock_names=DeviceRepository=MockDeviceRepository,CommandRepository=MockCommandRepository,EvaluationRuleRepository=MockEvaluationRuleRepository,TaskRepository=MockTaskRepository,ScheduledTaskRepository=MockScheduledTaskRepository

type Pagination = sharedUsecases.Pagination

var (
	ErrDeviceNotFound   = errors.New("device not found")
	ErrDeviceDuplicated = errors.New("device already exists")
	ErrCommandOverlap   = errors.New("command overlap detected")
)

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
