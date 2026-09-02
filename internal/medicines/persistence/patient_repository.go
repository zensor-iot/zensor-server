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

func NewPatientRepository(orm sql.ORM) (*SimplePatientRepository, error) {
	if err := orm.AutoMigrate(&internal.Patient{}); err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimplePatientRepository{orm: orm}, nil
}

var _ usecases.PatientRepository = (*SimplePatientRepository)(nil)

type SimplePatientRepository struct {
	orm sql.ORM
}

func (r *SimplePatientRepository) Create(ctx context.Context, patient medicinesDomain.Patient) error {
	entity := internal.FromPatient(patient)

	if err := r.orm.WithContext(ctx).Create(&entity).Error(); err != nil {
		return fmt.Errorf("creating patient in database: %w", err)
	}

	return nil
}

func (r *SimplePatientRepository) GetByID(ctx context.Context, id shareddomain.ID) (medicinesDomain.Patient, error) {
	var entity internal.Patient
	err := r.orm.WithContext(ctx).First(&entity, "id = ?", id.String()).Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return medicinesDomain.Patient{}, usecases.ErrPatientNotFound
	}
	if err != nil {
		return medicinesDomain.Patient{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimplePatientRepository) FindAllByTenant(
	ctx context.Context,
	tenantID shareddomain.ID,
	pagination usecases.Pagination,
) ([]medicinesDomain.Patient, int, error) {
	entities, total, err := paginateByFilter(ctx, r.orm, internal.Patient{},
		"tenant_id = ? AND deleted_at IS NULL", pagination, tenantID.String())
	if err != nil {
		return nil, 0, err
	}

	result := make([]medicinesDomain.Patient, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, total, nil
}

func (r *SimplePatientRepository) Update(ctx context.Context, patient medicinesDomain.Patient) error {
	entity := internal.FromPatient(patient)

	if err := r.orm.WithContext(ctx).Save(&entity).Error(); err != nil {
		return fmt.Errorf("updating patient in database: %w", err)
	}

	return nil
}

func (r *SimplePatientRepository) Delete(ctx context.Context, id shareddomain.ID) error {
	patient, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	patient.SoftDelete()

	return r.Update(ctx, patient)
}
