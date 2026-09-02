package usecases

//go:generate mockgen -source=repository_port.go -destination=../../../test/unit/doubles/medicines/usecases/repository_port_mock.go -package=usecases -mock_names=PatientRepository=MockPatientRepository,TreatmentRepository=MockTreatmentRepository,DoseRepository=MockDoseRepository

import (
	"context"
	"errors"
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

var (
	ErrPatientNotFound            = errors.New("patient not found")
	ErrTreatmentNotFound          = errors.New("treatment not found")
	ErrDoseNotFound               = errors.New("dose not found")
	ErrPatientDeleted             = errors.New("patient is deleted")
	ErrTreatmentDeleted           = errors.New("treatment is deleted")
	ErrPatientHasActiveTreatments = errors.New("patient has active treatments")
	ErrInvalidTreatmentSchedule   = errors.New("invalid treatment schedule")
	ErrTenantMismatch             = errors.New("treatment tenant does not match patient tenant")
	ErrAgendaWindowTooLarge       = errors.New("agenda window is too large")
)

type Pagination struct {
	Limit  int
	Offset int
}

// DoseWithContext carries the treatment and patient a dose belongs to, so
// callers rendering an agenda entry or a notification payload do not have to
// fetch them one by one.
type DoseWithContext struct {
	Dose      medicinesDomain.Dose
	Treatment medicinesDomain.Treatment
	Patient   medicinesDomain.Patient
}

type PatientRepository interface {
	Create(ctx context.Context, patient medicinesDomain.Patient) error
	GetByID(ctx context.Context, id shareddomain.ID) (medicinesDomain.Patient, error)
	FindAllByTenant(ctx context.Context, tenantID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Patient, int, error)
	Update(ctx context.Context, patient medicinesDomain.Patient) error
	Delete(ctx context.Context, id shareddomain.ID) error
}

type TreatmentRepository interface {
	Create(ctx context.Context, treatment medicinesDomain.Treatment) error
	GetByID(ctx context.Context, id shareddomain.ID) (medicinesDomain.Treatment, error)
	FindAllByPatient(ctx context.Context, patientID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Treatment, int, error)
	FindAllByTenant(ctx context.Context, tenantID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Treatment, int, error)
	FindAllActive(ctx context.Context) ([]medicinesDomain.Treatment, error)
	CountActiveByPatient(ctx context.Context, patientID shareddomain.ID) (int, error)
	Update(ctx context.Context, treatment medicinesDomain.Treatment) error
	Delete(ctx context.Context, id shareddomain.ID) error
}

type DoseRepository interface {
	CreateBatch(ctx context.Context, doses []medicinesDomain.Dose) error
	GetByID(ctx context.Context, id shareddomain.ID) (medicinesDomain.Dose, error)
	FindAllByTreatment(ctx context.Context, treatmentID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Dose, int, error)
	FindByTreatmentInWindow(ctx context.Context, treatmentID shareddomain.ID, from, to time.Time) ([]medicinesDomain.Dose, error)
	FindDueForReminder(ctx context.Context, from, to time.Time, limit int) ([]DoseWithContext, error)
	FindAgenda(ctx context.Context, tenantID shareddomain.ID, from, to time.Time, limit int) ([]DoseWithContext, error)
	Update(ctx context.Context, dose medicinesDomain.Dose) error
	// DeletePendingFrom removes the not yet resolved doses scheduled after the
	// given instant, so a schedule change does not leave stale future doses
	// behind. Resolved doses are the record of what happened and are untouched.
	DeletePendingFrom(ctx context.Context, treatmentID shareddomain.ID, from time.Time) error
}
