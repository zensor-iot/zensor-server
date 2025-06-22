package usecases

import (
	"context"
	"fmt"
	"zensor-server/internal/control_plane/domain"
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

func (s *SimpleScheduledTaskService) GetByID(ctx context.Context, id domain.ID) (domain.ScheduledTask, error) {
	scheduledTask, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return domain.ScheduledTask{}, fmt.Errorf("getting scheduled task by ID: %w", err)
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
