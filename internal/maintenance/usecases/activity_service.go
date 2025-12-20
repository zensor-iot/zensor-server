package usecases

//go:generate mockgen -source=./activity_service.go -destination=../../../test/unit/doubles/maintenance/usecases/maintenance_activity_service_mock.go -package=usecases -mock_names=ActivityService=MockActivityService

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type ActivityService interface {
	CreateActivity(ctx context.Context, activity maintenanceDomain.Activity) error
	GetActivity(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Activity, error)
	ListActivitiesByTenant(ctx context.Context, tenantID shareddomain.ID, pagination Pagination) ([]maintenanceDomain.Activity, int, error)
	UpdateActivity(ctx context.Context, activity maintenanceDomain.Activity) error
	DeleteActivity(ctx context.Context, id shareddomain.ID) error
	ActivateActivity(ctx context.Context, id shareddomain.ID) error
	DeactivateActivity(ctx context.Context, id shareddomain.ID) error
}

func NewActivityService(repository ActivityRepository) *SimpleActivityService {
	return &SimpleActivityService{
		repository: repository,
	}
}

var _ ActivityService = (*SimpleActivityService)(nil)

type SimpleActivityService struct {
	repository ActivityRepository
}

func (s *SimpleActivityService) CreateActivity(ctx context.Context, activity maintenanceDomain.Activity) error {
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

func (s *SimpleActivityService) GetActivity(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Activity, error) {
	activity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return maintenanceDomain.Activity{}, ErrActivityNotFound
		}
		slog.Error("getting maintenance activity", slog.String("error", err.Error()))
		return maintenanceDomain.Activity{}, fmt.Errorf("getting maintenance activity: %w", err)
	}

	return activity, nil
}

func (s *SimpleActivityService) ListActivitiesByTenant(
	ctx context.Context,
	tenantID shareddomain.ID,
	pagination Pagination,
) ([]maintenanceDomain.Activity, int, error) {
	activities, total, err := s.repository.FindAllByTenant(ctx, tenantID, pagination)
	if err != nil {
		slog.Error("listing maintenance activities", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("listing maintenance activities: %w", err)
	}

	return activities, total, nil
}

func (s *SimpleActivityService) UpdateActivity(ctx context.Context, activity maintenanceDomain.Activity) error {
	existingActivity, err := s.repository.GetByID(ctx, activity.ID)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return ErrActivityNotFound
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

func (s *SimpleActivityService) DeleteActivity(ctx context.Context, id shareddomain.ID) error {
	activity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return ErrActivityNotFound
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

func (s *SimpleActivityService) ActivateActivity(ctx context.Context, id shareddomain.ID) error {
	activity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return ErrActivityNotFound
		}
		return fmt.Errorf("getting maintenance activity: %w", err)
	}

	if activity.IsDeleted() {
		return errors.New("maintenance activity is deleted")
	}

	activity.Activate()
	return s.repository.Update(ctx, activity)
}

func (s *SimpleActivityService) DeactivateActivity(ctx context.Context, id shareddomain.ID) error {
	activity, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return ErrActivityNotFound
		}
		return fmt.Errorf("getting maintenance activity: %w", err)
	}

	if activity.IsDeleted() {
		return errors.New("maintenance activity is deleted")
	}

	activity.Deactivate()
	return s.repository.Update(ctx, activity)
}
