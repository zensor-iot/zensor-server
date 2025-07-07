package persistence

import (
	"context"
	"fmt"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
)

const (
	_tasksTopic    = "tasks"
	_commandsTopic = "device_commands"
)

func NewTaskRepository(
	publisherFactory pubsub.PublisherFactory,
	orm sql.ORM,
) (*SimpleTaskRepository, error) {
	taskPublisher, err := publisherFactory.New(_tasksTopic, &avro.AvroTask{})
	if err != nil {
		return nil, fmt.Errorf("creating task publisher: %w", err)
	}

	commandPublisher, err := publisherFactory.New(_commandsTopic, &avro.AvroCommand{})
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
	// Convert domain task to Avro task
	avroTask := &avro.AvroTask{
		ID:        task.ID.String(),
		DeviceID:  task.Device.ID.String(),
		Version:   int64(task.Version),
		CreatedAt: task.CreatedAt.Time,
		UpdatedAt: time.Now(),
	}

	if task.ScheduledTask != nil {
		scheduledTaskIDStr := task.ScheduledTask.ID.String()
		avroTask.ScheduledTaskID = &scheduledTaskIDStr
	}

	err := r.taskPublisher.Publish(ctx, pubsub.Key(task.ID), avroTask)
	if err != nil {
		return fmt.Errorf("publishing task to kafka: %w", err)
	}

	// Publish each command as Avro command
	for _, cmd := range task.Commands {
		avroCommand := &avro.AvroCommand{
			ID:            cmd.ID.String(),
			Version:       int(cmd.Version),
			DeviceID:      cmd.Device.ID.String(),
			DeviceName:    cmd.Device.Name,
			TaskID:        cmd.Task.ID.String(),
			PayloadIndex:  int(cmd.Payload.Index),
			PayloadValue:  int(cmd.Payload.Value),
			DispatchAfter: cmd.DispatchAfter.Time,
			Port:          int(cmd.Port),
			Priority:      string(cmd.Priority),
			CreatedAt:     time.Now(),
			Ready:         cmd.Ready,
			Sent:          cmd.Sent,
			SentAt:        cmd.SentAt.Time,
		}

		err := r.commandPublisher.Publish(ctx, pubsub.Key(cmd.ID), avroCommand)
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
		Order("created_at DESC").
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
		Where("scheduled_task_id = ?", scheduledTaskID.String()).
		Count(&total).
		Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.Task
	err = r.orm.
		WithContext(ctx).
		Where("scheduled_task_id = ?", scheduledTaskID.String()).
		Order("created_at DESC").
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
