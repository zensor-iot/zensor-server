package usecases

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/control_plane/domain"
)

var (
	ErrScheduledTaskNotFound = errors.New("scheduled task not found")
)

func NewScheduledTaskService(repository ScheduledTaskRepository) *SimpleScheduledTaskService {
	return &SimpleScheduledTaskService{
		repository: repository,
	}
}

var _ ScheduledTaskService = (*SimpleScheduledTaskService)(nil)

type SimpleScheduledTaskService struct {
	repository ScheduledTaskRepository
}

func (s *SimpleScheduledTaskService) Create(ctx context.Context, scheduledTask domain.ScheduledTask) error {
	err := s.repository.Create(ctx, scheduledTask)
	if err != nil {
		return fmt.Errorf("creating scheduled task: %w", err)
	}

	return nil
}

func (s *SimpleScheduledTaskService) FindAllByTenant(ctx context.Context, tenantID domain.ID) ([]domain.ScheduledTask, error) {
	scheduledTasks, err := s.repository.FindAllByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("finding scheduled tasks by tenant: %w", err)
	}

	return scheduledTasks, nil
}

func (s *SimpleScheduledTaskService) FindAllByTenantAndDevice(ctx context.Context, tenantID domain.ID, deviceID domain.ID, pagination Pagination) ([]domain.ScheduledTask, int, error) {
	scheduledTasks, total, err := s.repository.FindAllByTenantAndDevice(ctx, tenantID, deviceID, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("finding scheduled tasks by tenant and device: %w", err)
	}

	return scheduledTasks, total, nil
}

func (s *SimpleScheduledTaskService) GetByID(ctx context.Context, id domain.ID) (domain.ScheduledTask, error) {
	scheduledTask, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrScheduledTaskNotFound) {
			return domain.ScheduledTask{}, ErrScheduledTaskNotFound
		}
		return domain.ScheduledTask{}, err
	}

	return scheduledTask, nil
}

func (s *SimpleScheduledTaskService) Update(ctx context.Context, scheduledTask domain.ScheduledTask) error {
	err := s.repository.Update(ctx, scheduledTask)
	if err != nil {
		return fmt.Errorf("updating scheduled task: %w", err)
	}

	return nil
}

func (s *SimpleScheduledTaskService) Delete(ctx context.Context, id domain.ID) error {
	err := s.repository.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting scheduled task: %w", err)
	}

	return nil
}
