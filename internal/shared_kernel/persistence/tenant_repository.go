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

func NewTenantRepository(orm sql.ORM) (*SimpleTenantRepository, error) {
	err := orm.AutoMigrate(&internal.Tenant{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleTenantRepository{
		orm: orm,
	}, nil
}

var _ usecases.TenantRepository = (*SimpleTenantRepository)(nil)

type SimpleTenantRepository struct {
	orm sql.ORM
}

func (r *SimpleTenantRepository) Create(ctx context.Context, tenant domain.Tenant) error {
	entity := internal.FromTenant(tenant)

	err := r.orm.WithContext(ctx).Create(&entity).Error()
	if err != nil {
		return fmt.Errorf("creating tenant in database: %w", err)
	}

	return nil
}

func (r *SimpleTenantRepository) GetByID(ctx context.Context, id domain.ID) (domain.Tenant, error) {
	var entity internal.Tenant
	err := r.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id.String()).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.Tenant{}, usecases.ErrTenantNotFound
	}

	if err != nil {
		return domain.Tenant{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleTenantRepository) GetByName(ctx context.Context, name string) (domain.Tenant, error) {
	var entity internal.Tenant
	err := r.orm.
		WithContext(ctx).
		Where("name = ?", name).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.Tenant{}, usecases.ErrTenantNotFound
	}

	if err != nil {
		return domain.Tenant{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleTenantRepository) Update(ctx context.Context, tenant domain.Tenant) error {
	tenant.Version++
	tenant.UpdatedAt = time.Now()

	entity := internal.FromTenant(tenant)

	err := r.orm.WithContext(ctx).Save(&entity).Error()
	if err != nil {
		return fmt.Errorf("updating tenant in database: %w", err)
	}

	return nil
}

func (r *SimpleTenantRepository) FindAll(ctx context.Context, includeDeleted bool, pagination usecases.Pagination) ([]domain.Tenant, int, error) {
	var total int64
	query := r.orm.WithContext(ctx).Model(&internal.Tenant{})

	if !includeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	err := query.Count(&total).Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.Tenant
	query = r.orm.WithContext(ctx)

	if !includeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	err = query.Limit(pagination.Limit).Offset(pagination.Offset).Find(&entities).Error()
	if err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	result := make([]domain.Tenant, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, int(total), nil
}
