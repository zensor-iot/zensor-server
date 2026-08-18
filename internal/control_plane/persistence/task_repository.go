package persistence

import (
	"context"
	"fmt"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/domain"
)

func NewTaskRepository(orm sql.ORM) (*SimpleTaskRepository, error) {
	err := orm.AutoMigrate(&internal.Task{}, &internal.Command{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleTaskRepository{
		orm: orm,
	}, nil
}

var _ usecases.TaskRepository = (*SimpleTaskRepository)(nil)

type SimpleTaskRepository struct {
	orm sql.ORM
}

func (r *SimpleTaskRepository) Create(ctx context.Context, task domain.Task) error {
	return r.orm.WithContext(ctx).Transaction(func(tx sql.ORM) error {
		entity := internal.FromTask(task)
		if err := tx.Create(&entity).Error(); err != nil {
			return fmt.Errorf("creating task in database: %w", err)
		}

		for _, cmd := range task.Commands {
			cmdEntity := internal.FromCommand(cmd)
			if err := tx.Create(&cmdEntity).Error(); err != nil {
				return fmt.Errorf("creating command in database: %w", err)
			}
		}

		return nil
	})
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
