package usecases

import (
	"context"
	"errors"
	"log/slog"
	"zensor-server/internal/control_plane/domain"
)

var (
	errUnknown = errors.New("unknown error")
)

type DeviceService interface {
	CreateDevice(context.Context, domain.Device) error
}

func NewDeviceService(repository DeviceRepository) *SimpleDeviceService {
	return &SimpleDeviceService{
		repository,
	}
}

var _ DeviceService = &SimpleDeviceService{}

type SimpleDeviceService struct {
	repository DeviceRepository
}

func (s *SimpleDeviceService) CreateDevice(ctx context.Context, device domain.Device) error {
	err := s.repository.CreateDevice(ctx, device)
	if err != nil {
		slog.Error("creating device", slog.String("error", err.Error()))
		return errUnknown
	}

	return nil
}
