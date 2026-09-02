package usecases

//go:generate mockgen -source=./dose_service.go -destination=../../../test/unit/doubles/medicines/usecases/dose_service_mock.go -package=usecases -mock_names=DoseService=MockDoseService

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

const (
	_maxAgendaWindow = 31 * 24 * time.Hour
	_maxAgendaSize   = 500
)

type DoseService interface {
	GetDose(ctx context.Context, id shareddomain.ID) (medicinesDomain.Dose, error)
	ListDosesByTreatment(ctx context.Context, treatmentID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Dose, int, error)
	ListAgenda(ctx context.Context, tenantID shareddomain.ID, from, to time.Time) ([]DoseWithContext, error)
	AdministerDose(ctx context.Context, id shareddomain.ID, by medicinesDomain.ResolvedBy, notes *medicinesDomain.DoseNotes) error
	SkipDose(ctx context.Context, id shareddomain.ID, by medicinesDomain.ResolvedBy, notes *medicinesDomain.DoseNotes) error
}

func NewDoseService(repository DoseRepository) *SimpleDoseService {
	return &SimpleDoseService{repository: repository}
}

var _ DoseService = (*SimpleDoseService)(nil)

type SimpleDoseService struct {
	repository DoseRepository
}

func (s *SimpleDoseService) GetDose(ctx context.Context, id shareddomain.ID) (medicinesDomain.Dose, error) {
	dose, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return medicinesDomain.Dose{}, fmt.Errorf("getting dose: %w", err)
	}
	if dose.IsDeleted() {
		return medicinesDomain.Dose{}, ErrDoseNotFound
	}

	return dose, nil
}

func (s *SimpleDoseService) ListDosesByTreatment(ctx context.Context, treatmentID shareddomain.ID, pagination Pagination) ([]medicinesDomain.Dose, int, error) {
	doses, total, err := s.repository.FindAllByTreatment(ctx, treatmentID, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("listing doses: %w", err)
	}

	return doses, total, nil
}

func (s *SimpleDoseService) ListAgenda(ctx context.Context, tenantID shareddomain.ID, from, to time.Time) ([]DoseWithContext, error) {
	if to.Sub(from) > _maxAgendaWindow {
		return nil, ErrAgendaWindowTooLarge
	}

	entries, err := s.repository.FindAgenda(ctx, tenantID, from, to, _maxAgendaSize)
	if err != nil {
		return nil, fmt.Errorf("listing agenda: %w", err)
	}

	return entries, nil
}

// AdministerDose records that a dose was given. A dose scheduled a little ahead
// can still be administered: giving medicine slightly early is normal care, so
// unlike a maintenance execution it is not rejected.
func (s *SimpleDoseService) AdministerDose(ctx context.Context, id shareddomain.ID, by medicinesDomain.ResolvedBy, notes *medicinesDomain.DoseNotes) error {
	return s.resolve(ctx, id, func(dose *medicinesDomain.Dose) error {
		return dose.MarkAdministered(by, notes)
	})
}

func (s *SimpleDoseService) SkipDose(ctx context.Context, id shareddomain.ID, by medicinesDomain.ResolvedBy, notes *medicinesDomain.DoseNotes) error {
	return s.resolve(ctx, id, func(dose *medicinesDomain.Dose) error {
		return dose.MarkSkipped(by, notes)
	})
}

func (s *SimpleDoseService) resolve(
	ctx context.Context,
	id shareddomain.ID,
	apply func(*medicinesDomain.Dose) error,
) error {
	dose, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("getting dose: %w", err)
	}
	if dose.IsDeleted() {
		return ErrDoseNotFound
	}

	if err := apply(&dose); err != nil {
		return err
	}

	if err := s.repository.Update(ctx, dose); err != nil {
		return fmt.Errorf("updating dose: %w", err)
	}

	slog.Info("dose resolved",
		slog.String("id", dose.ID.String()),
		slog.String("status", string(dose.Status)))

	return nil
}
