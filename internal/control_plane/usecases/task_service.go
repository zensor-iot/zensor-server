package usecases

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/control_plane/domain"
)

func NewTaskService(repository TaskRepository) *SimpleTaskService {
	return &SimpleTaskService{repository: repository}
}

var _ TaskService = (*SimpleTaskService)(nil)

type SimpleTaskService struct {
	repository TaskRepository
}

func (s *SimpleTaskService) Create(ctx context.Context, task domain.Task) error {
	err := s.repository.Create(ctx, task)
	if err != nil {
		return fmt.Errorf("creating task: %w", err)
	}

	return nil
}

func (s *SimpleTaskService) Run(_ context.Context, task domain.Task) error {
	return errors.New("implement me")
}
