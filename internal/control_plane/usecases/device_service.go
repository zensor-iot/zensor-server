package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"
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

func (s *SimpleDeviceService) DevicesByTenant(ctx context.Context, tenantID domain.ID) ([]domain.Device, error) {
	devices, err := s.repository.FindByTenant(ctx, tenantID.String())
	if err != nil {
		slog.Error("getting devices by tenant",
			slog.String("tenant_id", tenantID.String()),
			slog.String("error", err.Error()))
		return nil, errUnknown
	}

	return devices, nil
}

func (s *SimpleDeviceService) UpdateDeviceDisplayName(ctx context.Context, deviceID domain.ID, displayName string) error {
	device, err := s.repository.Get(ctx, deviceID.String())
	if errors.Is(err, ErrDeviceNotFound) {
		slog.Warn("device not found for display name update", slog.String("device_id", deviceID.String()))
		return ErrDeviceNotFound
	}
	if err != nil {
		slog.Error("getting device for display name update", slog.String("error", err.Error()))
		return errUnknown
	}

	device.UpdateDisplayName(displayName)

	err = s.repository.UpdateDevice(ctx, device)
	if err != nil {
		slog.Error("updating device display name", slog.String("error", err.Error()))
		return errUnknown
	}

	slog.Info("device display name updated successfully",
		slog.String("device_id", deviceID.String()),
		slog.String("display_name", displayName))

	return nil
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

func (s *SimpleDeviceService) UpdateLastMessageReceivedAt(ctx context.Context, deviceName string) error {
	device, err := s.repository.FindByName(ctx, deviceName)
	if errors.Is(err, ErrDeviceNotFound) {
		slog.Warn("device not found for message update", slog.String("device_name", deviceName))
		return ErrDeviceNotFound
	}
	if err != nil {
		slog.Error("getting device for message update", slog.String("error", err.Error()))
		return fmt.Errorf("getting device: %w", err)
	}

	device.UpdateLastMessageReceivedAt(utils.Time{Time: time.Now()})

	err = s.repository.UpdateDevice(ctx, device)
	if err != nil {
		slog.Error("updating device last message timestamp", slog.String("error", err.Error()))
		return fmt.Errorf("updating device: %w", err)
	}

	slog.Debug("device last message timestamp updated",
		slog.String("device_name", deviceName),
		slog.String("device_id", device.ID.String()))

	return nil
}
