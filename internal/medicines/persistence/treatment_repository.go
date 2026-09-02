package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/medicines/persistence/internal"
	"zensor-server/internal/medicines/usecases"

	medicinesDomain "zensor-server/internal/medicines/domain"

	shareddomain "zensor-server/internal/shared_kernel/domain"
)

func NewTreatmentRepository(orm sql.ORM) (*SimpleTreatmentRepository, error) {
	if err := orm.AutoMigrate(&internal.Treatment{}); err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleTreatmentRepository{orm: orm}, nil
}

var _ usecases.TreatmentRepository = (*SimpleTreatmentRepository)(nil)

type SimpleTreatmentRepository struct {
	orm sql.ORM
}

func (r *SimpleTreatmentRepository) Create(ctx context.Context, treatment medicinesDomain.Treatment) error {
	entity := internal.FromTreatment(treatment)

	if err := r.orm.WithContext(ctx).Create(&entity).Error(); err != nil {
		return fmt.Errorf("creating treatment in database: %w", err)
	}

	return nil
}

func (r *SimpleTreatmentRepository) GetByID(ctx context.Context, id shareddomain.ID) (medicinesDomain.Treatment, error) {
	var entity internal.Treatment
	err := r.orm.WithContext(ctx).First(&entity, "id = ?", id.String()).Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return medicinesDomain.Treatment{}, usecases.ErrTreatmentNotFound
	}
	if err != nil {
		return medicinesDomain.Treatment{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleTreatmentRepository) FindAllByPatient(
	ctx context.Context,
	patientID shareddomain.ID,
	pagination usecases.Pagination,
) ([]medicinesDomain.Treatment, int, error) {
	return r.paginate(ctx, "patient_id = ? AND deleted_at IS NULL", pagination, patientID.String())
}

func (r *SimpleTreatmentRepository) FindAllByTenant(
	ctx context.Context,
	tenantID shareddomain.ID,
	pagination usecases.Pagination,
) ([]medicinesDomain.Treatment, int, error) {
	return r.paginate(ctx, "tenant_id = ? AND deleted_at IS NULL", pagination, tenantID.String())
}

func (r *SimpleTreatmentRepository) FindAllActive(ctx context.Context) ([]medicinesDomain.Treatment, error) {
	var entities []internal.Treatment
	err := r.orm.
		WithContext(ctx).
		Where("is_active = ? AND deleted_at IS NULL", true).
		Find(&entities).
		Error()
	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	result := make([]medicinesDomain.Treatment, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, nil
}

func (r *SimpleTreatmentRepository) CountActiveByPatient(ctx context.Context, patientID shareddomain.ID) (int, error) {
	var total int64
	err := r.orm.
		WithContext(ctx).
		Model(&internal.Treatment{}).
		Where("patient_id = ? AND is_active = ? AND deleted_at IS NULL", patientID.String(), true).
		Count(&total).
		Error()
	if err != nil {
		return 0, fmt.Errorf("count query: %w", err)
	}

	return int(total), nil
}

func (r *SimpleTreatmentRepository) Update(ctx context.Context, treatment medicinesDomain.Treatment) error {
	entity := internal.FromTreatment(treatment)

	if err := r.orm.WithContext(ctx).Save(&entity).Error(); err != nil {
		return fmt.Errorf("updating treatment in database: %w", err)
	}

	return nil
}

func (r *SimpleTreatmentRepository) Delete(ctx context.Context, id shareddomain.ID) error {
	treatment, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	treatment.SoftDelete()

	return r.Update(ctx, treatment)
}

func (r *SimpleTreatmentRepository) paginate(
	ctx context.Context,
	filter string,
	pagination usecases.Pagination,
	args ...any,
) ([]medicinesDomain.Treatment, int, error) {
	entities, total, err := paginateByFilter(ctx, r.orm, internal.Treatment{}, filter, pagination, args...)
	if err != nil {
		return nil, 0, err
	}

	result := make([]medicinesDomain.Treatment, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, total, nil
}
