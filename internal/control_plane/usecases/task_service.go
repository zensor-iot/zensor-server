package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"zensor-server/internal/shared_kernel/domain"
)

func NewTaskService(repository TaskRepository, commandRepository CommandRepository, deviceRepository DeviceRepository) *SimpleTaskService {
	return &SimpleTaskService{
		repository:        repository,
		commandRepository: commandRepository,
		deviceRepository:  deviceRepository,
	}
}

var _ TaskService = (*SimpleTaskService)(nil)

type SimpleTaskService struct {
	repository        TaskRepository
	commandRepository CommandRepository
	deviceRepository  DeviceRepository
}

func (s *SimpleTaskService) Create(ctx context.Context, task domain.Task) error {
	// Validate for command overlaps before creating the task
	err := s.validateCommandOverlaps(ctx, task)
	if err != nil {
		return err
	}

	err = s.repository.Create(ctx, task)
	if err != nil {
		return fmt.Errorf("creating task: %w", err)
	}

	return nil
}

func (s *SimpleTaskService) validateCommandOverlaps(ctx context.Context, newTask domain.Task) error {
	pendingCommands, err := s.commandRepository.FindPendingByDevice(ctx, newTask.Device.ID)
	if err != nil {
		return fmt.Errorf("finding pending commands: %w", err)
	}

	for _, newCmd := range newTask.Commands {
		if slices.ContainsFunc(pendingCommands, newCmd.OverlapsWith) {
			return ErrCommandOverlap
		}
	}

	return nil
}

func (s *SimpleTaskService) Run(_ context.Context, task domain.Task) error {
	return errors.New("implement me")
}

func (s *SimpleTaskService) FindAllByDevice(ctx context.Context, deviceID domain.ID, pagination Pagination) ([]domain.Task, int, error) {
	device, err := s.deviceRepository.Get(ctx, string(deviceID))
	if err != nil {
		return nil, 0, fmt.Errorf("finding device: %w", err)
	}

	tasks, total, err := s.repository.FindAllByDevice(ctx, device, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("finding tasks by device: %w", err)
	}

	// Load commands for each task
	for i := range tasks {
		commands, err := s.commandRepository.FindByTaskID(ctx, tasks[i].ID)
		if err != nil {
			return nil, 0, fmt.Errorf("finding commands for task %s: %w", tasks[i].ID, err)
		}
		tasks[i].Commands = commands
	}

	return tasks, total, nil
}

func (s *SimpleTaskService) FindAllByScheduledTask(ctx context.Context, scheduledTaskID domain.ID, pagination Pagination) ([]domain.Task, int, error) {
	tasks, total, err := s.repository.FindAllByScheduledTask(ctx, scheduledTaskID, pagination)
	if err != nil {
		return nil, 0, fmt.Errorf("finding tasks by scheduled task: %w", err)
	}

	// Load commands for each task
	for i := range tasks {
		commands, err := s.commandRepository.FindByTaskID(ctx, tasks[i].ID)
		if err != nil {
			return nil, 0, fmt.Errorf("finding commands for task %s: %w", tasks[i].ID, err)
		}
		tasks[i].Commands = commands
	}

	return tasks, total, nil
}
