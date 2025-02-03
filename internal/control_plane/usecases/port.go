package usecases

import (
	"context"
	"zensor-server/internal/control_plane/domain"
)

type DeviceRepository interface {
	CreateDevice(context.Context, domain.Device) error
	FindAll(context.Context) ([]domain.Device, error)
}
