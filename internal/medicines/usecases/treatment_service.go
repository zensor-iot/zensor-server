package usecases

//go:generate mockgen -source=./treatment_service.go -destination=../../../test/unit/doubles/medicines/usecases/treatment_service_mock.go -package=usecases -mock_names=TreatmentService=MockTreatmentService

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type TreatmentService interface {
	CreateTreatment(ctx context.Context, treatment medicinesDomain.Treatment) error
	GetTreatment(ctx context.Context, id shareddomain.ID) (medicinesDomain.Treatment, error)
	ListTreatmentsByPatient(ctx context.Context, patientID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Treatment, int, error)
	ListTreatmentsByTenant(ctx context.Context, tenantID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Treatment, int, error)
	UpdateTreatment(ctx context.Context, treatment medicinesDomain.Treatment) error
	DeleteTreatment(ctx context.Context, id shareddomain.ID) error
	ActivateTreatment(ctx context.Context, id shareddomain.ID) error
	DeactivateTreatment(ctx context.Context, id shareddomain.ID) error
}

func NewTreatmentService(
	repository TreatmentRepository,
	patientRepository PatientRepository,
	doseRepository DoseRepository,
) *SimpleTreatmentService {
	return &SimpleTreatmentService{
		repository:        repository,
		patientRepository: patientRepository,
		doseRepository:    doseRepository,
	}
}

var _ TreatmentService = (*SimpleTreatmentService)(nil)

type SimpleTreatmentService struct {
	repository        TreatmentRepository
	patientRepository PatientRepository
	doseRepository    DoseRepository
}

func (s *SimpleTreatmentService) CreateTreatment(ctx context.Context, treatment medicinesDomain.Treatment) error {
	if err := s.validateAgainstPatient(ctx, treatment); err != nil {
		return err
	}

	// Validating here as well as in the builder keeps a schedule that cannot
	// produce doses from ever reaching the database, where it would make the
	// worker fail on every tick from then on.
	if err := treatment.Schedule.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTreatmentSchedule, err)
	}

	if err := s.repository.Create(ctx, treatment); err != nil {
		return fmt.Errorf("creating treatment: %w", err)
	}

	slog.Info("treatment created",
		slog.String("id", treatment.ID.String()),
		slog.String("patient_id", treatment.PatientID.String()))

	return nil
}

func (s *SimpleTreatmentService) GetTreatment(ctx context.Context, id shareddomain.ID) (medicinesDomain.Treatment, error) {
	treatment, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return medicinesDomain.Treatment{}, fmt.Errorf("getting treatment: %w", err)
	}
	if treatment.IsDeleted() {
		return medicinesDomain.Treatment{}, ErrTreatmentNotFound
	}

	return treatment, nil
}

func (s *SimpleTreatmentService) ListTreatmentsByPatient(ctx context.Context, patientID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Treatment, int, error) {
	treatments, total, err := s.repository.FindAllByPatient(ctx, patientID, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("listing treatments by patient: %w", err)
	}

	return treatments, total, nil
}

func (s *SimpleTreatmentService) ListTreatmentsByTenant(ctx context.Context, tenantID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Treatment, int, error) {
	treatments, total, err := s.repository.FindAllByTenant(ctx, tenantID, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("listing treatments by tenant: %w", err)
	}

	return treatments, total, nil
}

func (s *SimpleTreatmentService) UpdateTreatment(ctx context.Context, treatment medicinesDomain.Treatment) error {
	existing, err := s.repository.GetByID(ctx, treatment.ID)
	if err != nil {
		return fmt.Errorf("getting treatment: %w", err)
	}
	if existing.IsDeleted() {
		return ErrTreatmentDeleted
	}

	if err := treatment.Schedule.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidTreatmentSchedule, err)
	}

	if err := s.repository.Update(ctx, treatment); err != nil {
		return fmt.Errorf("updating treatment: %w", err)
	}

	return s.discardFutureDoses(ctx, treatment.ID)
}

func (s *SimpleTreatmentService) DeleteTreatment(ctx context.Context, id shareddomain.ID) error {
	treatment, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("getting treatment: %w", err)
	}
	if treatment.IsDeleted() {
		return ErrTreatmentNotFound
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting treatment: %w", err)
	}

	return s.discardFutureDoses(ctx, id)
}

func (s *SimpleTreatmentService) ActivateTreatment(ctx context.Context, id shareddomain.ID) error {
	treatment, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("getting treatment: %w", err)
	}
	if treatment.IsDeleted() {
		return ErrTreatmentNotFound
	}

	treatment.Activate()
	if err := s.repository.Update(ctx, treatment); err != nil {
		return fmt.Errorf("activating treatment: %w", err)
	}

	return nil
}

func (s *SimpleTreatmentService) DeactivateTreatment(ctx context.Context, id shareddomain.ID) error {
	treatment, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("getting treatment: %w", err)
	}
	if treatment.IsDeleted() {
		return ErrTreatmentNotFound
	}

	treatment.Deactivate()
	if err := s.repository.Update(ctx, treatment); err != nil {
		return fmt.Errorf("deactivating treatment: %w", err)
	}

	return s.discardFutureDoses(ctx, id)
}

func (s *SimpleTreatmentService) validateAgainstPatient(ctx context.Context, treatment medicinesDomain.Treatment) error {
	patient, err := s.patientRepository.GetByID(ctx, treatment.PatientID)
	if err != nil {
		return fmt.Errorf("getting patient: %w", err)
	}
	if patient.IsDeleted() {
		return ErrPatientDeleted
	}
	if patient.TenantID != treatment.TenantID {
		return ErrTenantMismatch
	}

	return nil
}

// discardFutureDoses drops the pending doses that the previous schedule had
// already materialised. The worker rebuilds them from the current definition on
// its next tick, and the resolved ones stay as the record of what happened.
func (s *SimpleTreatmentService) discardFutureDoses(ctx context.Context, treatmentID shareddomain.ID) error {
	if err := s.doseRepository.DeletePendingFrom(ctx, treatmentID, time.Now()); err != nil {
		return fmt.Errorf("discarding future doses: %w", err)
	}

	return nil
}
