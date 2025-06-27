package persistence

import (
	"context"
	"fmt"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/sql"
)

func NewCommandRepository(orm sql.ORM) (*SimpleCommandRepository, error) {
	err := orm.AutoMigrate(&internal.Command{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating command: %w", err)
	}

	return &SimpleCommandRepository{
		orm: orm,
	}, nil
}

var _ usecases.CommandRepository = (*SimpleCommandRepository)(nil)

type SimpleCommandRepository struct {
	orm sql.ORM
}

func (r *SimpleCommandRepository) FindAllPending(ctx context.Context) ([]domain.Command, error) {
	var entities internal.CommandSet
	err := r.orm.
		WithContext(ctx).
		Where("sent = ?", false).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	return entities.ToDomain(), nil
}

func (r *SimpleCommandRepository) FindPendingByDevice(ctx context.Context, deviceID domain.ID) ([]domain.Command, error) {
	var entities internal.CommandSet
	err := r.orm.
		WithContext(ctx).
		Where("sent = ? AND device_id = ?", false, deviceID.String()).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	return entities.ToDomain(), nil
}

func (r *SimpleCommandRepository) FindByTaskID(ctx context.Context, taskID domain.ID) ([]domain.Command, error) {
	var entities internal.CommandSet
	err := r.orm.
		WithContext(ctx).
		Where("task_id = ?", taskID.String()).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	return entities.ToDomain(), nil
}
