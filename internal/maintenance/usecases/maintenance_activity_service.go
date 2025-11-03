package usecases

//go:generate mockgen -source=maintenance_activity_service.go -destination=../../../test/unit/doubles/maintenance/usecases/maintenance_activity_service_mock.go -package=usecases -mock_names=MaintenanceActivityService=MockMaintenanceActivityService

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type MaintenanceActivityService interface {
	CreateActivity(ctx context.Context, activity maintenanceDomain.MaintenanceActivity) error
	GetActivity(ctx context.Context, id shareddomain.ID) (maintenanceDomain.MaintenanceActivity, error)
	ListActivitiesByTenant(ctx context.Context, tenantID shareddomain.ID, pagination Pagination) ([]maintenanceDomain.MaintenanceActivity, int, error)
	UpdateActivity(ctx context.Context, activity maintenanceDomain.MaintenanceActivity) error
	DeleteActivity(ctx context.Context, id shareddomain.ID) error
	ActivateActivity(ctx context.Context, id shareddomain.ID) error
	DeactivateActivity(ctx context.Context, id shareddomain.ID) error
}

func NewMaintenanceActivityService(repository MaintenanceActivityRepository) *SimpleMaintenanceActivityService {
	return &SimpleMaintenanceActivityService{
		repository: repository,
	}
}

var _ MaintenanceActivityService = (*SimpleMaintenanceActivityService)(nil)

type SimpleMaintenanceActivityService struct {
	repository MaintenanceActivityRepository
}

func (s *SimpleMaintenanceActivityService) CreateActivity(ctx context.Context, activity maintenanceDomain.MaintenanceActivity) error {
	err := s.repository.Create(ctx, activity)
	if err != nil {
		slog.Error("creating maintenance activity", slog.String("error", err.Error()))
		return fmt.Errorf("creating maintenance activity: %w", err)
	}

	slog.Info("maintenance activity created successfully",
		slog.String("id", activity.ID.String()),
		slog.String("tenant_id", activity.TenantID.String()))

	return nil
}

func (s *SimpleMaintenanceActivityService) GetActivity(ctx context.Context, id shareddomain.ID) (maintenanceDomain.MaintenanceActivity, error) {
	activity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrMaintenanceActivityNotFound) {
			return maintenanceDomain.MaintenanceActivity{}, ErrMaintenanceActivityNotFound
		}
		slog.Error("getting maintenance activity", slog.String("error", err.Error()))
		return maintenanceDomain.MaintenanceActivity{}, fmt.Errorf("getting maintenance activity: %w", err)
	}

	return activity, nil
}

func (s *SimpleMaintenanceActivityService) ListActivitiesByTenant(
	ctx context.Context,
	tenantID shareddomain.ID,
	pagination Pagination,
) ([]maintenanceDomain.MaintenanceActivity, int, error) {
	activities, total, err := s.repository.FindAllByTenant(ctx, tenantID, pagination)
	if err != nil {
		slog.Error("listing maintenance activities", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("listing maintenance activities: %w", err)
	}

	return activities, total, nil
}

func (s *SimpleMaintenanceActivityService) UpdateActivity(ctx context.Context, activity maintenanceDomain.MaintenanceActivity) error {
	existingActivity, err := s.repository.GetByID(ctx, activity.ID)
	if err != nil {
		if errors.Is(err, ErrMaintenanceActivityNotFound) {
			return ErrMaintenanceActivityNotFound
		}
		return fmt.Errorf("getting maintenance activity: %w", err)
	}

	if existingActivity.IsDeleted() {
		return errors.New("maintenance activity is deleted")
	}

	err = s.repository.Update(ctx, activity)
	if err != nil {
		slog.Error("updating maintenance activity", slog.String("error", err.Error()))
		return fmt.Errorf("updating maintenance activity: %w", err)
	}

	return nil
}

func (s *SimpleMaintenanceActivityService) DeleteActivity(ctx context.Context, id shareddomain.ID) error {
	activity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrMaintenanceActivityNotFound) {
			return ErrMaintenanceActivityNotFound
		}
		return fmt.Errorf("getting maintenance activity: %w", err)
	}

	if activity.IsDeleted() {
		return errors.New("maintenance activity is already deleted")
	}

	err = s.repository.Delete(ctx, id)
	if err != nil {
		slog.Error("deleting maintenance activity", slog.String("error", err.Error()))
		return fmt.Errorf("deleting maintenance activity: %w", err)
	}

	return nil
}

func (s *SimpleMaintenanceActivityService) ActivateActivity(ctx context.Context, id shareddomain.ID) error {
	activity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrMaintenanceActivityNotFound) {
			return ErrMaintenanceActivityNotFound
		}
		return fmt.Errorf("getting maintenance activity: %w", err)
	}

	if activity.IsDeleted() {
		return errors.New("maintenance activity is deleted")
	}

	activity.Activate()
	return s.repository.Update(ctx, activity)
}

func (s *SimpleMaintenanceActivityService) DeactivateActivity(ctx context.Context, id shareddomain.ID) error {
	activity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrMaintenanceActivityNotFound) {
			return ErrMaintenanceActivityNotFound
		}
		return fmt.Errorf("getting maintenance activity: %w", err)
	}

	if activity.IsDeleted() {
		return errors.New("maintenance activity is deleted")
	}

	activity.Deactivate()
	return s.repository.Update(ctx, activity)
}
