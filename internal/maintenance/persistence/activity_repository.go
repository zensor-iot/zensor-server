package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/infra/sql"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	"zensor-server/internal/maintenance/persistence/internal"
	"zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

func NewActivityRepository(orm sql.ORM) (*SimpleActivityRepository, error) {
	err := orm.AutoMigrate(&internal.Activity{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleActivityRepository{
		orm: orm,
	}, nil
}

var _ usecases.ActivityRepository = (*SimpleActivityRepository)(nil)

type SimpleActivityRepository struct {
	orm sql.ORM
}

func (r *SimpleActivityRepository) Create(ctx context.Context, activity maintenanceDomain.Activity) error {
	entity := internal.FromActivity(activity)

	err := r.orm.WithContext(ctx).Create(&entity).Error()
	if err != nil {
		return fmt.Errorf("creating maintenance activity in database: %w", err)
	}

	return nil
}

func (r *SimpleActivityRepository) GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Activity, error) {
	var entity internal.Activity
	err := r.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id.String()).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return maintenanceDomain.Activity{}, usecases.ErrActivityNotFound
	}

	if err != nil {
		return maintenanceDomain.Activity{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleActivityRepository) FindAllByTenant(
	ctx context.Context,
	tenantID shareddomain.ID,
	pagination usecases.Pagination,
) ([]maintenanceDomain.Activity, int, error) {
	var total int64
	query := r.orm.WithContext(ctx).Model(&internal.Activity{})

	err := query.Where("tenant_id = ? AND deleted_at IS NULL", tenantID.String()).Count(&total).Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.Activity
	err = query.
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID.String()).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&entities).
		Error()

	if err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	result := make([]maintenanceDomain.Activity, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, int(total), nil
}

func (r *SimpleActivityRepository) FindAllActive(ctx context.Context) ([]maintenanceDomain.Activity, error) {
	var entities []internal.Activity
	err := r.orm.
		WithContext(ctx).
		Where("is_active = ? AND deleted_at IS NULL", true).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	result := make([]maintenanceDomain.Activity, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, nil
}

func (r *SimpleActivityRepository) Update(ctx context.Context, activity maintenanceDomain.Activity) error {
	entity := internal.FromActivity(activity)

	err := r.orm.WithContext(ctx).Save(&entity).Error()
	if err != nil {
		return fmt.Errorf("updating maintenance activity in database: %w", err)
	}

	return nil
}

func (r *SimpleActivityRepository) Delete(ctx context.Context, id shareddomain.ID) error {
	activity, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	activity.SoftDelete()
	return r.Update(ctx, activity)
}
