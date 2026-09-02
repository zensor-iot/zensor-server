// Package usecases provides the business logic of the medicines module.
package usecases

//go:generate mockgen -source=./patient_service.go -destination=../../../test/unit/doubles/medicines/usecases/patient_service_mock.go -package=usecases -mock_names=PatientService=MockPatientService

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	medicinesDomain "zensor-server/internal/medicines/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"
)

type PatientService interface {
	CreatePatient(ctx context.Context, patient medicinesDomain.Patient) error
	GetPatient(ctx context.Context, id shareddomain.ID) (medicinesDomain.Patient, error)
	ListPatientsByTenant(ctx context.Context, tenantID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Patient, int, error)
	UpdatePatient(ctx context.Context, patient medicinesDomain.Patient) error
	DeletePatient(ctx context.Context, id shareddomain.ID) error
}

func NewPatientService(
	repository PatientRepository,
	treatmentRepository TreatmentRepository,
	tenantService sharedUsecases.TenantService,
) *SimplePatientService {
	return &SimplePatientService{
		repository:          repository,
		treatmentRepository: treatmentRepository,
		tenantService:       tenantService,
	}
}

var _ PatientService = (*SimplePatientService)(nil)

type SimplePatientService struct {
	repository          PatientRepository
	treatmentRepository TreatmentRepository
	tenantService       sharedUsecases.TenantService
}

func (s *SimplePatientService) CreatePatient(ctx context.Context, patient medicinesDomain.Patient) error {
	if _, err := s.tenantService.GetTenant(ctx, patient.TenantID); err != nil {
		if errors.Is(err, sharedUsecases.ErrTenantNotFound) {
			slog.Error("tenant not found when creating patient",
				slog.String("tenant_id", patient.TenantID.String()))
			return fmt.Errorf("tenant not found: %w", err)
		}
		return fmt.Errorf("getting tenant: %w", err)
	}

	if err := s.repository.Create(ctx, patient); err != nil {
		return fmt.Errorf("creating patient: %w", err)
	}

	slog.Info("patient created",
		slog.String("id", patient.ID.String()),
		slog.String("tenant_id", patient.TenantID.String()))

	return nil
}

func (s *SimplePatientService) GetPatient(ctx context.Context, id shareddomain.ID) (medicinesDomain.Patient, error) {
	patient, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return medicinesDomain.Patient{}, fmt.Errorf("getting patient: %w", err)
	}
	if patient.IsDeleted() {
		return medicinesDomain.Patient{}, ErrPatientNotFound
	}

	return patient, nil
}

func (s *SimplePatientService) ListPatientsByTenant(ctx context.Context, tenantID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Patient, int, error) {
	patients, total, err := s.repository.FindAllByTenant(ctx, tenantID, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("listing patients: %w", err)
	}

	return patients, total, nil
}

func (s *SimplePatientService) UpdatePatient(ctx context.Context, patient medicinesDomain.Patient) error {
	existing, err := s.repository.GetByID(ctx, patient.ID)
	if err != nil {
		return fmt.Errorf("getting patient: %w", err)
	}
	if existing.IsDeleted() {
		return ErrPatientDeleted
	}

	if err := s.repository.Update(ctx, patient); err != nil {
		return fmt.Errorf("updating patient: %w", err)
	}

	return nil
}

// DeletePatient refuses to remove a patient that is still under treatment, so
// an active prescription can never be left without the person it belongs to.
func (s *SimplePatientService) DeletePatient(ctx context.Context, id shareddomain.ID) error {
	patient, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("getting patient: %w", err)
	}
	if patient.IsDeleted() {
		return ErrPatientNotFound
	}

	active, err := s.treatmentRepository.CountActiveByPatient(ctx, id)
	if err != nil {
		return fmt.Errorf("counting active treatments: %w", err)
	}
	if active > 0 {
		return ErrPatientHasActiveTreatments
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting patient: %w", err)
	}

	slog.Info("patient deleted", slog.String("id", id.String()))

	return nil
}
