package persistence

import (
	"context"
	"fmt"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
)

const (
	_tasksTopic    = "tasks"
	_commandsTopic = "device_commands"
)

func NewTaskRepository(
	publisherFactory pubsub.PublisherFactory,
	orm sql.ORM,
) (*SimpleTaskRepository, error) {
	taskPublisher, err := publisherFactory.New(_tasksTopic, internal.Task{})
	if err != nil {
		return nil, fmt.Errorf("creating task publisher: %w", err)
	}

	commandPublisher, err := publisherFactory.New(_commandsTopic, internal.Task{})
	if err != nil {
		return nil, fmt.Errorf("creating command publisher: %w", err)
	}

	err = orm.AutoMigrate(&internal.Task{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleTaskRepository{
		taskPublisher:    taskPublisher,
		commandPublisher: commandPublisher,
		orm:              orm,
	}, nil
}

var _ usecases.TaskRepository = (*SimpleTaskRepository)(nil)

type SimpleTaskRepository struct {
	taskPublisher    pubsub.Publisher
	commandPublisher pubsub.Publisher
	orm              sql.ORM
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

func (r *SimpleTaskRepository) FindAllByDevice(ctx context.Context, device domain.Device, pagination usecases.Pagination) ([]domain.Task, int, error) {
	var total int64
	err := r.orm.
		WithContext(ctx).
		Model(&internal.Task{}).
		Where("device_id = ?", device.ID.String()).
		Count(&total).
		Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.Task
	err = r.orm.
		WithContext(ctx).
		Where("device_id = ?", device.ID.String()).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&entities).
		Error()
	if err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	tasks := make([]domain.Task, len(entities))
	for i, entity := range entities {
		tasks[i] = entity.ToDomain()
		tasks[i].Device = device
	}

	return tasks, int(total), nil
}

func (r *SimpleTaskRepository) FindAllByScheduledTask(ctx context.Context, scheduledTaskID domain.ID, pagination usecases.Pagination) ([]domain.Task, int, error) {
	var total int64
	err := r.orm.
		WithContext(ctx).
		Model(&internal.Task{}).
		Where("scheduled_task LIKE ?", "%"+scheduledTaskID.String()+"%").
		Count(&total).
		Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.Task
	err = r.orm.
		WithContext(ctx).
		Where("scheduled_task LIKE ?", "%"+scheduledTaskID.String()+"%").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&entities).
		Error()
	if err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	tasks := make([]domain.Task, len(entities))
	for i, entity := range entities {
		tasks[i] = entity.ToDomain()
	}

	return tasks, int(total), nil
}
