package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/kafka"
	"zensor-server/internal/infra/sql"
)

var (
	ErrDeviceNotFound   = errors.New("device not found")
	ErrDeviceDuplicated = errors.New("device already exists")
)

func NewDeviceRepository(publisherFactory *kafka.KafkaPublisherFactory, orm sql.ORM) (*SimpleDeviceRepository, error) {
	publisher, err := publisherFactory.CreatePublisher("devices", internal.Device{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	return &SimpleDeviceRepository{
		publisher: publisher,
		orm:       orm,
	}, nil
}

var _ usecases.DeviceRepository = &SimpleDeviceRepository{}

type SimpleDeviceRepository struct {
	publisher kafka.KafkaPublisher
	orm       sql.ORM
}

func (s *SimpleDeviceRepository) CreateDevice(ctx context.Context, device domain.Device) error {
	data := internal.FromDevice(device)
	currentDevice, err := s.GetDeviceByName(ctx, device.Name)
	if err != nil && err != ErrDeviceNotFound {
		return fmt.Errorf("getting device: %w", err)
	}

	if currentDevice != (domain.Device{}) {
		return ErrDeviceDuplicated
	}

	err = s.publisher.Publish(ctx, device.ID, data)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (s *SimpleDeviceRepository) GetDeviceByName(ctx context.Context, name string) (domain.Device, error) {
	var entity internal.Device
	err := s.orm.
		WithContext(ctx).
		Where("name = ?", name).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.Device{}, ErrDeviceNotFound
	}

	if err != nil {
		return domain.Device{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}
