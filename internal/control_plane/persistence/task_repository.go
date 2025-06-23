package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
)

const (
	_tasksTopic    = "tasks"
	_commandsTopic = "device_commands"
)

func NewTaskRepository(
	publisherFactory pubsub.PublisherFactory,
) (*SimpleTaskRepository, error) {
	taskPublisher, err := publisherFactory.New(_tasksTopic, internal.Task{})
	if err != nil {
		return nil, fmt.Errorf("creating task publisher: %w", err)
	}

	commandPublisher, err := publisherFactory.New(_commandsTopic, internal.Task{})
	if err != nil {
		return nil, fmt.Errorf("creating command publisher: %w", err)
	}

	return &SimpleTaskRepository{
		taskPublisher:    taskPublisher,
		commandPublisher: commandPublisher,
	}, nil
}

var _ usecases.TaskRepository = (*SimpleTaskRepository)(nil)

type SimpleTaskRepository struct {
	taskPublisher    pubsub.Publisher
	commandPublisher pubsub.Publisher
}

func (r *SimpleTaskRepository) Create(ctx context.Context, task domain.Task) error {
	data := internal.FromTask(task)
	err := r.taskPublisher.Publish(ctx, pubsub.Key(task.ID), data)
	if err != nil {
		return fmt.Errorf("publishing task to kafka: %w", err)
	}

	for _, cmd := range task.Commands {
		data := internal.FromCommand(cmd)
		data.TaskID = string(task.ID)
		err := r.commandPublisher.Publish(ctx, pubsub.Key(cmd.ID), data)
		if err != nil {
			return fmt.Errorf("publishing command to kafka: %w", err)
		}
	}

	return nil
}

func (r *SimpleTaskRepository) FindAllByDevice(ctx context.Context, device domain.Device) ([]domain.Task, error) {
	return nil, errors.New("implement me")
}
