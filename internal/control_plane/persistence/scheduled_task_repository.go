package persistence

import (
	"context"
	"errors"
	"fmt"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
	"zensor-server/internal/shared_kernel/domain"
)

const (
	_scheduledTasksTopic = "scheduled_tasks"
)

func NewScheduledTaskRepository(publisherFactory pubsub.PublisherFactory, orm sql.ORM) (*SimpleScheduledTaskRepository, error) {
	publisher, err := publisherFactory.New(_scheduledTasksTopic, &avro.AvroScheduledTask{})
	if err != nil {
		return nil, fmt.Errorf("creating scheduled task publisher: %w", err)
	}

	err = orm.AutoMigrate(&internal.ScheduledTask{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleScheduledTaskRepository{
		publisher: publisher,
		orm:       orm,
	}, nil
}

var _ usecases.ScheduledTaskRepository = (*SimpleScheduledTaskRepository)(nil)

type SimpleScheduledTaskRepository struct {
	publisher pubsub.Publisher
	orm       sql.ORM
}

func (r *SimpleScheduledTaskRepository) Create(ctx context.Context, scheduledTask domain.ScheduledTask) error {
	avroScheduledTask := avro.ToAvroScheduledTask(scheduledTask)

	err := r.publisher.Publish(ctx, pubsub.Key(scheduledTask.ID), avroScheduledTask)
	if err != nil {
		return fmt.Errorf("publishing scheduled task to kafka: %w", err)
	}

	return nil
}

func (r *SimpleScheduledTaskRepository) FindAllByTenant(ctx context.Context, tenantID domain.ID) ([]domain.ScheduledTask, error) {
	var entities []internal.ScheduledTask
	err := r.orm.
		WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID.String()).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	result := make([]domain.ScheduledTask, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, nil
}

func (r *SimpleScheduledTaskRepository) FindAllByTenantAndDevice(ctx context.Context, tenantID domain.ID, deviceID domain.ID, pagination usecases.Pagination) ([]domain.ScheduledTask, int, error) {
	var entities []internal.ScheduledTask
	var total int64

	// Get total count
	err := r.orm.
		WithContext(ctx).
		Model(&internal.ScheduledTask{}).
		Where("tenant_id = ? AND device_id = ? AND deleted_at IS NULL", tenantID.String(), deviceID.String()).
		Count(&total).
		Error()

	if err != nil {
		return nil, 0, fmt.Errorf("counting scheduled tasks: %w", err)
	}

	// Get paginated results
	err = r.orm.
		WithContext(ctx).
		Where("tenant_id = ? AND device_id = ? AND deleted_at IS NULL", tenantID.String(), deviceID.String()).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&entities).
		Error()

	if err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	result := make([]domain.ScheduledTask, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, int(total), nil
}

func (r *SimpleScheduledTaskRepository) FindAllActive(ctx context.Context) ([]domain.ScheduledTask, error) {
	var entities []internal.ScheduledTask
	err := r.orm.
		WithContext(ctx).
		Where("is_active = ? AND deleted_at IS NULL", true).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	result := make([]domain.ScheduledTask, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, nil
}

func (r *SimpleScheduledTaskRepository) Update(ctx context.Context, scheduledTask domain.ScheduledTask) error {
	avroScheduledTask := avro.ToAvroScheduledTask(scheduledTask)

	err := r.publisher.Publish(ctx, pubsub.Key(scheduledTask.ID), avroScheduledTask)
	if err != nil {
		return fmt.Errorf("publishing scheduled task update to kafka: %w", err)
	}

	return nil
}

func (r *SimpleScheduledTaskRepository) Delete(ctx context.Context, id domain.ID) error {
	scheduledTask, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("getting scheduled task for deletion: %w", err)
	}

	scheduledTask.SoftDelete()

	avroScheduledTask := avro.ToAvroScheduledTask(scheduledTask)

	err = r.publisher.Publish(ctx, pubsub.Key(scheduledTask.ID), avroScheduledTask)
	if err != nil {
		return fmt.Errorf("publishing scheduled task deletion to kafka: %w", err)
	}

	return nil
}

func (r *SimpleScheduledTaskRepository) GetByID(ctx context.Context, id domain.ID) (domain.ScheduledTask, error) {
	var entity internal.ScheduledTask
	err := r.orm.
		WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id.String()).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return domain.ScheduledTask{}, usecases.ErrScheduledTaskNotFound
	}

	if err != nil {
		return domain.ScheduledTask{}, err
	}

	return entity.ToDomain(), nil
}
