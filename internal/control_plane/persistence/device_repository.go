package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/domain"
)

func NewDeviceRepository(orm sql.ORM) (*SimpleDeviceRepository, error) {
	err := orm.AutoMigrate(&internal.Device{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleDeviceRepository{
		orm: orm,
	}, nil
}

var _ usecases.DeviceRepository = (*SimpleDeviceRepository)(nil)

type SimpleDeviceRepository struct {
	orm sql.ORM
}

func (s *SimpleDeviceRepository) CreateDevice(ctx context.Context, device domain.Device) error {
	currentDevice, err := s.GetByName(ctx, device.Name)
	if err != nil && err != usecases.ErrDeviceNotFound {
		return fmt.Errorf("getting device: %w", err)
	}

	if currentDevice.ID != "" {
		return usecases.ErrDeviceDuplicated
	}

	entity := internal.FromDevice(device)
	err = s.orm.WithContext(ctx).Create(&entity).Error()
	if err != nil {
		return fmt.Errorf("creating device in database: %w", err)
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

	entity := internal.FromDevice(device)
	err = s.orm.WithContext(ctx).Save(&entity).Error()
	if err != nil {
		return fmt.Errorf("updating device in database: %w", err)
	}

	return nil
}

func (s *SimpleDeviceRepository) AddEvaluationRule(ctx context.Context, device domain.Device, rule domain.EvaluationRule) error {
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

func (s *SimpleDeviceRepository) FindAll(ctx context.Context, pagination usecases.Pagination) ([]domain.Device, int, error) {
	var total int64
	err := s.orm.
		WithContext(ctx).
		Model(&internal.Device{}).
		Count(&total).
		Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.Device
	err = s.orm.
		WithContext(ctx).
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
