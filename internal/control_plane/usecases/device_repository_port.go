package usecases

import (
	"context"
	"errors"
	"zensor-server/internal/control_plane/domain"
)

var (
	ErrDeviceNotFound   = errors.New("device not found")
	ErrDeviceDuplicated = errors.New("device already exists")
)

type DeviceRepository interface {
	CreateDevice(context.Context, domain.Device) error
	UpdateDevice(context.Context, domain.Device) error
	Get(context.Context, string) (domain.Device, error)
	FindAll(context.Context) ([]domain.Device, error)
	AddEvaluationRule(context.Context, domain.Device, domain.EvaluationRule) error
	FindAllEvaluationRules(context.Context, domain.Device) ([]domain.EvaluationRule, error)
}

type CommandRepository interface {
	FindAllPending(context.Context) ([]domain.Command, error)
}
