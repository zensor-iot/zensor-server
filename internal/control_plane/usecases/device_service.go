package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"zensor-server/internal/control_plane/domain"
)

var (
	errUnknown = errors.New("unknown error")
)

type DeviceService interface {
	CreateDevice(context.Context, domain.Device) error
	AllDevices(context.Context) ([]domain.Device, error)
	QueueCommand(context.Context, domain.Command) error
}

func NewDeviceService(
	repository DeviceRepository,
	publisher CommandPublisher,
) *SimpleDeviceService {
	return &SimpleDeviceService{
		repository,
		publisher,
	}
}

var _ DeviceService = &SimpleDeviceService{}

type SimpleDeviceService struct {
	repository DeviceRepository
	publisher  CommandPublisher
}

func (s *SimpleDeviceService) CreateDevice(ctx context.Context, device domain.Device) error {
	err := s.repository.CreateDevice(ctx, device)
	if err != nil {
		slog.Error("creating device", slog.String("error", err.Error()))
		return errUnknown
	}

	return nil
}

func (s *SimpleDeviceService) AllDevices(ctx context.Context) ([]domain.Device, error) {
	devices, err := s.repository.FindAll(ctx)
	if err != nil {
		slog.Error("getting all devices", slog.String("error", err.Error()))
		return nil, errUnknown
	}

	return devices, nil
}

func (s *SimpleDeviceService) QueueCommand(ctx context.Context, cmd domain.Command) error {
	device, err := s.repository.Get(ctx, cmd.Device.ID)
	if errors.Is(err, ErrDeviceNotFound) {
		return ErrDeviceNotFound
	}
	if err != nil {
		return fmt.Errorf("get device: %w", err)
	}

	cmd.Device = device
	err = s.publisher.Dispatch(ctx, cmd)
	if err != nil {
		return fmt.Errorf("dispatch event: %w", err)
	}

	return nil
}
