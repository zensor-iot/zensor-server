package usecases

//go:generate mockgen -source=maintenance_execution_service.go -destination=../../../test/unit/doubles/maintenance/usecases/maintenance_execution_service_mock.go -package=usecases -mock_names=ExecutionService=MockExecutionService

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type ExecutionService interface {
	CreateExecution(ctx context.Context, execution maintenanceDomain.Execution) error
	GetExecution(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Execution, error)
	ListExecutionsByActivity(ctx context.Context, activityID shareddomain.ID, pagination Pagination) ([]maintenanceDomain.Execution, int, error)
	MarkExecutionCompleted(ctx context.Context, id shareddomain.ID, completedBy string) error
}

func NewExecutionService(
	repository ExecutionRepository,
	activityRepository ActivityRepository,
) *SimpleExecutionService {
	return &SimpleExecutionService{
		repository:         repository,
		activityRepository: activityRepository,
	}
}

var _ ExecutionService = (*SimpleExecutionService)(nil)

type SimpleExecutionService struct {
	repository         ExecutionRepository
	activityRepository ActivityRepository
}

func (s *SimpleExecutionService) CreateExecution(ctx context.Context, execution maintenanceDomain.Execution) error {
	_, err := s.activityRepository.GetByID(ctx, execution.ActivityID)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return errors.New("activity not found")
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

func (s *SimpleExecutionService) GetExecution(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Execution, error) {
	execution, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return maintenanceDomain.Execution{}, ErrExecutionNotFound
		}
		slog.Error("getting maintenance execution", slog.String("error", err.Error()))
		return maintenanceDomain.Execution{}, fmt.Errorf("getting maintenance execution: %w", err)
	}

	return execution, nil
}

func (s *SimpleExecutionService) ListExecutionsByActivity(
	ctx context.Context,
	activityID shareddomain.ID,
	pagination Pagination,
) ([]maintenanceDomain.Execution, int, error) {
	executions, total, err := s.repository.FindAllByActivity(ctx, activityID, pagination)
	if err != nil {
		slog.Error("listing maintenance executions", slog.String("error", err.Error()))
		return nil, 0, fmt.Errorf("listing maintenance executions: %w", err)
	}

	return executions, total, nil
}

func (s *SimpleExecutionService) MarkExecutionCompleted(ctx context.Context, id shareddomain.ID, completedBy string) error {
	execution, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return ErrExecutionNotFound
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
