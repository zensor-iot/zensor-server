package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
)

func NewDeviceRepository(publisherFactory pubsub.PublisherFactory, orm sql.ORM) (*SimpleDeviceRepository, error) {
	publisher, err := publisherFactory.New("devices", internal.Device{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	return &SimpleDeviceRepository{
		publisher: publisher,
		orm:       orm,
	}, nil
}

var _ usecases.DeviceRepository = (*SimpleDeviceRepository)(nil)

type SimpleDeviceRepository struct {
	publisher pubsub.Publisher
	orm       sql.ORM
}

func (s *SimpleDeviceRepository) CreateDevice(ctx context.Context, device domain.Device) error {
	data := internal.FromDevice(device)
	currentDevice, err := s.GetByName(ctx, device.Name)
	if err != nil && err != usecases.ErrDeviceNotFound {
		return fmt.Errorf("getting device: %w", err)
	}

	if currentDevice.ID != "" {
		return usecases.ErrDeviceDuplicated
	}

	err = s.publisher.Publish(ctx, pubsub.Key(device.ID), data)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (s *SimpleDeviceRepository) UpdateDevice(ctx context.Context, device domain.Device) error {
	data := internal.FromDevice(device)
	currentDevice, err := s.GetByName(ctx, device.Name)
	if err != nil && err != usecases.ErrDeviceNotFound {
		return fmt.Errorf("getting device: %w", err)
	}

	if currentDevice.ID != "" {
		return usecases.ErrDeviceDuplicated
	}

	err = s.publisher.Publish(ctx, pubsub.Key(device.ID), data)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (s *SimpleDeviceRepository) AddEvaluationRule(ctx context.Context, device domain.Device, rule domain.EvaluationRule) error {
	// deviceData := internal.FromDevice(device)
	// evaluationRuleData := internal.FromEvaluationRule(rule)
	// currentDevice, err := s.GetByName(ctx, device.Name)
	// if err != nil && err != usecases.ErrDeviceNotFound {
	// 	return fmt.Errorf("getting device: %w", err)
	// }

	// if currentDevice.ID != "" {
	// 	return usecases.ErrDeviceDuplicated
	// }

	// err = s.publisher.Publish(ctx, pubsub.Key(device.ID), data)
	// if err != nil {
	// 	return fmt.Errorf("publishing to kafka: %w", err)
	// }

	return nil
}

func (s *SimpleDeviceRepository) GetByName(ctx context.Context, name string) (domain.Device, error) {
	var entity internal.Device
	err := s.orm.
		WithContext(ctx).
		Where("name = ?", name).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.Device{}, usecases.ErrDeviceNotFound
	}

	if err != nil {
		return domain.Device{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (s *SimpleDeviceRepository) Get(ctx context.Context, id string) (domain.Device, error) {
	var entity internal.Device
	err := s.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.Device{}, usecases.ErrDeviceNotFound
	}

	if err != nil {
		return domain.Device{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (s *SimpleDeviceRepository) FindAll(ctx context.Context) ([]domain.Device, error) {
	var entities []internal.Device
	err := s.orm.
		WithContext(ctx).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	result := make([]domain.Device, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, nil
}

func (s *SimpleDeviceRepository) FindAllEvaluationRules(ctx context.Context, device domain.Device) ([]domain.EvaluationRule, error) {
	var entities []internal.EvaluationRule
	err := s.orm.
		WithContext(ctx).
		Where("device_id = ?", device.ID).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	result := make([]domain.EvaluationRule, len(entities))
	// for i, entity := range entities {
	// 	result[i] = entity.ToDomain()
	// }

	return result, nil
}

var _ usecases.CommandRepository = (*SimpleDeviceRepository)(nil)

func (r *SimpleDeviceRepository) FindAllPending(ctx context.Context) ([]domain.Command, error) {
	var entities internal.CommandSet
	err := r.orm.
		WithContext(ctx).
		Where("sent = ?", false).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	return entities.ToDomain(), nil
}
