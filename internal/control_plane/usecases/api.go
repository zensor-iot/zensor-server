package usecases

import (
	"context"
	"zensor-server/internal/control_plane/domain"
)

//go:generate mockgen -source=./api.go -destination=../../../test/unit/doubles/control_plane/usecases/api.go

type DeviceService interface {
	CreateDevice(context.Context, domain.Device) error
	GetDevice(context.Context, domain.ID) (domain.Device, error)
	AllDevices(context.Context) ([]domain.Device, error)
	DevicesByTenant(context.Context, domain.ID) ([]domain.Device, error)
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
	Run(context.Context, domain.Task) error
}
