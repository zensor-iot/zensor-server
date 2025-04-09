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
	QueueCommand(context.Context, domain.Command) error
}

type EvaluationRuleService interface {
	Create(context.Context, domain.EvaluationRule) error
	Get(context.Context, domain.ID) (domain.EvaluationRule, error)
	FindAllByDevice(context.Context, domain.Device) ([]domain.EvaluationRule, error)
	Update(context.Context, domain.EvaluationRule) error
	Delete(context.Context, domain.ID) error
}
