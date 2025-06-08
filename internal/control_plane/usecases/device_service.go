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
	if errors.Is(ErrDeviceDuplicated, err) {
		slog.Warn("device duplicated", slog.String("name", device.Name))
		return ErrDeviceDuplicated
	}

	if err != nil {
		slog.Error("creating device", slog.String("error", err.Error()))
		return errUnknown
	}

	return nil
}

func (s *SimpleDeviceService) GetDevice(ctx context.Context, id domain.ID) (domain.Device, error) {
	devices, err := s.repository.Get(ctx, string(id))
	if err != nil {
		slog.Error("getting device", slog.String("error", err.Error()))
		return domain.Device{}, errUnknown
	}

	return devices, nil
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
	if cmd.Port == 0 {
		cmd.Port = 1
	}

	device, err := s.repository.Get(ctx, cmd.Device.ID.String())
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

func (s *SimpleDeviceService) AdoptDeviceToTenant(ctx context.Context, tenantID, deviceID domain.ID) error {
	device, err := s.repository.Get(ctx, deviceID.String())
	if errors.Is(err, ErrDeviceNotFound) {
		slog.Warn("device not found for adoption", slog.String("device_id", deviceID.String()))
		return ErrDeviceNotFound
	}
	if err != nil {
		slog.Error("getting device for adoption", slog.String("error", err.Error()))
		return fmt.Errorf("getting device: %w", err)
	}

	device.AdoptToTenant(tenantID)

	err = s.repository.UpdateDevice(ctx, device)
	if err != nil {
		slog.Error("updating device for adoption", slog.String("error", err.Error()))
		return fmt.Errorf("updating device: %w", err)
	}

	slog.Info("device adopted to tenant successfully",
		slog.String("device_id", deviceID.String()),
		slog.String("tenant_id", tenantID.String()))

	return nil
}

func (s *SimpleDeviceService) QueueCommandSequence(ctx context.Context, cmd domain.CommandSequence) error {
	for _, command := range cmd.Commands {
		err := s.QueueCommand(ctx, command)
		if err != nil {
			return err
		}
	}

	return nil
}
