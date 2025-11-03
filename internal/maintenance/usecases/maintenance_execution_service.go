package usecases

//go:generate mockgen -source=maintenance_execution_service.go -destination=../../../test/unit/doubles/maintenance/usecases/maintenance_execution_service_mock.go -package=usecases -mock_names=MaintenanceExecutionService=MockMaintenanceExecutionService

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type MaintenanceExecutionService interface {
	CreateExecution(ctx context.Context, execution maintenanceDomain.MaintenanceExecution) error
	GetExecution(ctx context.Context, id shareddomain.ID) (maintenanceDomain.MaintenanceExecution, error)
	ListExecutionsByActivity(ctx context.Context, activityID shareddomain.ID, pagination Pagination) ([]maintenanceDomain.MaintenanceExecution, int, error)
	MarkExecutionCompleted(ctx context.Context, id shareddomain.ID, completedBy string) error
}

func NewMaintenanceExecutionService(
	repository MaintenanceExecutionRepository,
	activityRepository MaintenanceActivityRepository,
) *SimpleMaintenanceExecutionService {
	return &SimpleMaintenanceExecutionService{
		repository:         repository,
		activityRepository: activityRepository,
	}
}

var _ MaintenanceExecutionService = (*SimpleMaintenanceExecutionService)(nil)

type SimpleMaintenanceExecutionService struct {
	repository         MaintenanceExecutionRepository
	activityRepository MaintenanceActivityRepository
}

func (s *SimpleMaintenanceExecutionService) CreateExecution(ctx context.Context, execution maintenanceDomain.MaintenanceExecution) error {
	_, err := s.activityRepository.GetByID(ctx, execution.ActivityID)
	if err != nil {
		if errors.Is(err, ErrMaintenanceActivityNotFound) {
			return errors.New("maintenance activity not found")
		}
		return fmt.Errorf("getting maintenance activity: %w", err)
	}

	err = s.repository.Create(ctx, execution)
	if err != nil {
		slog.Error("creating maintenance execution", slog.String("error", err.Error()))
		return fmt.Errorf("creating maintenance execution: %w", err)
	}

	slog.Info("maintenance execution created successfully",
		slog.String("id", execution.ID.String()),
		slog.String("activity_id", execution.ActivityID.String()))

	return nil
}

func (s *SimpleMaintenanceExecutionService) GetExecution(ctx context.Context, id shareddomain.ID) (maintenanceDomain.MaintenanceExecution, error) {
	execution, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrMaintenanceExecutionNotFound) {
			return maintenanceDomain.MaintenanceExecution{}, ErrMaintenanceExecutionNotFound
		}
		slog.Error("getting maintenance execution", slog.String("error", err.Error()))
		return maintenanceDomain.MaintenanceExecution{}, fmt.Errorf("getting maintenance execution: %w", err)
	}

	return execution, nil
}

func (s *SimpleMaintenanceExecutionService) ListExecutionsByActivity(
	ctx context.Context,
	activityID shareddomain.ID,
	pagination Pagination,
) ([]maintenanceDomain.MaintenanceExecution, int, error) {
	executions, total, err := s.repository.FindAllByActivity(ctx, activityID, pagination)
	if err != nil {
		slog.Error("listing maintenance executions", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("listing maintenance executions: %w", err)
	}

	return executions, total, nil
}

func (s *SimpleMaintenanceExecutionService) MarkExecutionCompleted(ctx context.Context, id shareddomain.ID, completedBy string) error {
	execution, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrMaintenanceExecutionNotFound) {
			return ErrMaintenanceExecutionNotFound
		}
		return fmt.Errorf("getting maintenance execution: %w", err)
	}

	if execution.IsDeleted() {
		return errors.New("maintenance execution is deleted")
	}

	if execution.IsCompleted() {
		return errors.New("maintenance execution is already completed")
	}

	err = s.repository.MarkCompleted(ctx, id, completedBy)
	if err != nil {
		slog.Error("marking maintenance execution as completed", slog.String("error", err.Error()))
		return fmt.Errorf("marking maintenance execution as completed: %w", err)
	}

	return nil
}
