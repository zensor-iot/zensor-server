package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
)

func NewDeviceRepository(publisherFactory pubsub.PublisherFactory, orm sql.ORM) (*SimpleDeviceRepository, error) {
	publisher, err := publisherFactory.New("devices", &avro.AvroDevice{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	err = orm.AutoMigrate(&internal.Device{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
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
	currentDevice, err := s.GetByName(ctx, device.Name)
	if err != nil && err != usecases.ErrDeviceNotFound {
		return fmt.Errorf("getting device: %w", err)
	}

	if currentDevice.ID != "" {
		return usecases.ErrDeviceDuplicated
	}

	// Convert domain device to Avro device
	avroDevice := &avro.AvroDevice{
		ID:          device.ID.String(),
		Version:     1, // Default version for new devices
		Name:        device.Name,
		DisplayName: device.DisplayName,
		AppEUI:      device.AppEUI,
		DevEUI:      device.DevEUI,
		AppKey:      device.AppKey,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if device.TenantID != nil {
		tenantIDStr := device.TenantID.String()
		avroDevice.TenantID = &tenantIDStr
	}

	if !device.LastMessageReceivedAt.IsZero() {
		lastMessageStr := device.LastMessageReceivedAt.Time
		avroDevice.LastMessageReceivedAt = &lastMessageStr
	}

	err = s.publisher.Publish(ctx, pubsub.Key(device.ID), avroDevice)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (s *SimpleDeviceRepository) UpdateDevice(ctx context.Context, device domain.Device) error {
	currentDevice, err := s.GetByName(ctx, device.Name)
	if err != nil && err != usecases.ErrDeviceNotFound {
		return fmt.Errorf("getting device: %w", err)
	}

	if currentDevice.ID == "" {
		return usecases.ErrDeviceNotFound
	}

	// Convert domain device to Avro device
	avroDevice := &avro.AvroDevice{
		ID:          device.ID.String(),
		Version:     1, // Default version for updates
		Name:        device.Name,
		DisplayName: device.DisplayName,
		AppEUI:      device.AppEUI,
		DevEUI:      device.DevEUI,
		AppKey:      device.AppKey,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if device.TenantID != nil {
		tenantIDStr := device.TenantID.String()
		avroDevice.TenantID = &tenantIDStr
	}

	if !device.LastMessageReceivedAt.IsZero() {
		lastMessageStr := device.LastMessageReceivedAt.Time
		avroDevice.LastMessageReceivedAt = &lastMessageStr
	}

	err = s.publisher.Publish(ctx, pubsub.Key(device.ID), avroDevice)
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

func (s *SimpleDeviceRepository) FindByName(ctx context.Context, name string) (domain.Device, error) {
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

func (s *SimpleDeviceRepository) GetByName(ctx context.Context, name string) (domain.Device, error) {
	return s.FindByName(ctx, name)
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

func (s *SimpleDeviceRepository) FindByTenant(ctx context.Context, tenantID string, pagination usecases.Pagination) ([]domain.Device, int, error) {
	var total int64
	err := s.orm.
		WithContext(ctx).
		Model(&internal.Device{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).
		Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.Device
	err = s.orm.
		WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&entities).
		Error()
	if err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	result := make([]domain.Device, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, int(total), nil
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
