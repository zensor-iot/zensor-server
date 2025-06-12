package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"zensor-server/internal/control_plane/domain"
)

func NewTaskService(repository TaskRepository, commandRepository CommandRepository) *SimpleTaskService {
	return &SimpleTaskService{
		repository:        repository,
		commandRepository: commandRepository,
	}
}

var _ TaskService = (*SimpleTaskService)(nil)

type SimpleTaskService struct {
	repository        TaskRepository
	commandRepository CommandRepository
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
