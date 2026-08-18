package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/persistence/internal"
	"zensor-server/internal/shared_kernel/usecases"
)

func NewTenantConfigurationRepository(orm sql.ORM) (*SimpleTenantConfigurationRepository, error) {
	err := orm.AutoMigrate(&internal.TenantConfiguration{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleTenantConfigurationRepository{
		orm: orm,
	}, nil
}

var _ usecases.TenantConfigurationRepository = (*SimpleTenantConfigurationRepository)(nil)

type SimpleTenantConfigurationRepository struct {
	orm sql.ORM
}

func (r *SimpleTenantConfigurationRepository) Create(ctx context.Context, config domain.TenantConfiguration) error {
	internalConfig := internal.FromTenantConfiguration(config)
	err := r.orm.WithContext(ctx).Create(&internalConfig).Error()
	if err != nil {
		return fmt.Errorf("creating tenant configuration in database: %w", err)
	}

	return nil
}

func (r *SimpleTenantConfigurationRepository) GetByTenantID(ctx context.Context, tenantID domain.ID) (domain.TenantConfiguration, error) {
	var entity internal.TenantConfiguration
	err := r.orm.
		WithContext(ctx).
		Where("tenant_id = ?", tenantID.String()).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.TenantConfiguration{}, usecases.ErrTenantConfigurationNotFound
	}

	if err != nil {
		return domain.TenantConfiguration{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleTenantConfigurationRepository) Update(ctx context.Context, config domain.TenantConfiguration) error {
	config.Version++
	config.UpdatedAt = time.Now()

	internalConfig := internal.FromTenantConfiguration(config)
	err := r.orm.WithContext(ctx).Save(&internalConfig).Error()
	if err != nil {
		return fmt.Errorf("updating tenant configuration in database: %w", err)
	}

	return nil
}
