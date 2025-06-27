package usecases

import (
	"context"
	"errors"
	"zensor-server/internal/control_plane/domain"
)

var (
	ErrDeviceNotFound   = errors.New("device not found")
	ErrDeviceDuplicated = errors.New("device already exists")
	ErrCommandOverlap   = errors.New("command overlap detected")
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
	FindAll(context.Context) ([]domain.Device, error)
	FindByTenant(context.Context, string, Pagination) ([]domain.Device, int, error)
	AddEvaluationRule(context.Context, domain.Device, domain.EvaluationRule) error
	FindAllEvaluationRules(context.Context, domain.Device) ([]domain.EvaluationRule, error)
}

type CommandRepository interface {
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
	FindAllActive(context.Context) ([]domain.ScheduledTask, error)
	Update(context.Context, domain.ScheduledTask) error
	GetByID(context.Context, domain.ID) (domain.ScheduledTask, error)
}
