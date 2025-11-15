package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	"zensor-server/internal/maintenance/persistence/internal"
	"zensor-server/internal/maintenance/usecases"
	"zensor-server/internal/shared_kernel/avro"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

func NewMaintenanceExecutionRepository(
	publisherFactory pubsub.PublisherFactory,
	orm sql.ORM,
) (*SimpleMaintenanceExecutionRepository, error) {
	publisher, err := publisherFactory.New(_maintenanceExecutionsTopic, &avro.AvroMaintenanceExecution{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	err = orm.AutoMigrate(&internal.MaintenanceExecution{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleMaintenanceExecutionRepository{
		publisher: publisher,
		orm:       orm,
	}, nil
}

var _ usecases.MaintenanceExecutionRepository = (*SimpleMaintenanceExecutionRepository)(nil)

type SimpleMaintenanceExecutionRepository struct {
	publisher pubsub.Publisher
	orm       sql.ORM
}

func (r *SimpleMaintenanceExecutionRepository) Create(ctx context.Context, execution maintenanceDomain.MaintenanceExecution) error {
	entity := internal.FromMaintenanceExecution(execution)

	err := r.orm.WithContext(ctx).Create(&entity).Error()
	if err != nil {
		return fmt.Errorf("creating maintenance execution in database: %w", err)
	}

	avroExecution := convertToAvroMaintenanceExecution(execution)

	slog.Debug("publishing maintenance execution to pubsub", slog.String("execution_id", execution.ID.String()))
	err = r.publisher.Publish(ctx, pubsub.Key(execution.ID), avroExecution)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}
	slog.Debug("published maintenance execution to pubsub", slog.String("execution_id", execution.ID.String()))

	return nil
}

func (r *SimpleMaintenanceExecutionRepository) GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.MaintenanceExecution, error) {
	var entity internal.MaintenanceExecution
	err := r.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id.String()).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return maintenanceDomain.MaintenanceExecution{}, usecases.ErrMaintenanceExecutionNotFound
	}

	if err != nil {
		return maintenanceDomain.MaintenanceExecution{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleMaintenanceExecutionRepository) FindAllByActivity(
	ctx context.Context,
	activityID shareddomain.ID,
	pagination usecases.Pagination,
) ([]maintenanceDomain.MaintenanceExecution, int, error) {
	var total int64
	query := r.orm.WithContext(ctx).Model(&internal.MaintenanceExecution{})

	err := query.Where("activity_id = ? AND deleted_at IS NULL", activityID.String()).Count(&total).Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.MaintenanceExecution
	err = query.
		Where("activity_id = ? AND deleted_at IS NULL", activityID.String()).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&entities).
		Error()

	if err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	result := make([]maintenanceDomain.MaintenanceExecution, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, int(total), nil
}

func (r *SimpleMaintenanceExecutionRepository) Update(ctx context.Context, execution maintenanceDomain.MaintenanceExecution) error {
	entity := internal.FromMaintenanceExecution(execution)

	err := r.orm.WithContext(ctx).Save(&entity).Error()
	if err != nil {
		return fmt.Errorf("updating maintenance execution in database: %w", err)
	}

	avroExecution := convertToAvroMaintenanceExecution(execution)

	err = r.publisher.Publish(ctx, pubsub.Key(execution.ID), avroExecution)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (r *SimpleMaintenanceExecutionRepository) MarkCompleted(ctx context.Context, id shareddomain.ID, completedBy string) error {
	execution, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	execution.MarkCompleted(completedBy)
	return r.Update(ctx, execution)
}

func (r *SimpleMaintenanceExecutionRepository) FindAllOverdue(ctx context.Context, tenantID shareddomain.ID) ([]maintenanceDomain.MaintenanceExecution, error) {
	var entities []internal.MaintenanceExecution
	now := time.Now()

	err := r.orm.
		WithContext(ctx).
		Joins("JOIN maintenance_activities ON maintenance_executions.activity_id = maintenance_activities.id").
		Where("maintenance_activities.tenant_id = ? AND maintenance_executions.deleted_at IS NULL AND maintenance_executions.completed_at IS NULL AND maintenance_executions.scheduled_date < ?", tenantID.String(), now).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	result := make([]maintenanceDomain.MaintenanceExecution, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, nil
}

func (r *SimpleMaintenanceExecutionRepository) FindAllDueSoon(ctx context.Context, tenantID shareddomain.ID, days int) ([]maintenanceDomain.MaintenanceExecution, error) {
	var entities []internal.MaintenanceExecution
	now := time.Now()
	dueDate := now.AddDate(0, 0, days)

	err := r.orm.
		WithContext(ctx).
		Joins("JOIN maintenance_activities ON maintenance_executions.activity_id = maintenance_activities.id").
		Where("maintenance_activities.tenant_id = ? AND maintenance_executions.deleted_at IS NULL AND maintenance_executions.completed_at IS NULL AND maintenance_executions.scheduled_date BETWEEN ? AND ?", tenantID.String(), now, dueDate).
		Find(&entities).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	result := make([]maintenanceDomain.MaintenanceExecution, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, nil
}

func convertToAvroMaintenanceExecution(execution maintenanceDomain.MaintenanceExecution) *avro.AvroMaintenanceExecution {
	fieldValues := make(map[string]string)
	for k, v := range execution.FieldValues {
		if str, err := json.Marshal(v); err == nil {
			fieldValues[k] = string(str)
		}
	}

	result := &avro.AvroMaintenanceExecution{
		ID:            execution.ID.String(),
		Version:       int(execution.Version),
		ActivityID:    execution.ActivityID.String(),
		ScheduledDate: execution.ScheduledDate.Time,
		OverdueDays:   int(execution.OverdueDays),
		FieldValues:   fieldValues,
		CreatedAt:     execution.CreatedAt.Time,
		UpdatedAt:     execution.UpdatedAt.Time,
	}

	if execution.CompletedAt != nil {
		result.CompletedAt = &execution.CompletedAt.Time
	}

	if execution.CompletedBy != nil {
		completedBy := string(*execution.CompletedBy)
		result.CompletedBy = &completedBy
	}

	if execution.DeletedAt != nil {
		result.DeletedAt = &execution.DeletedAt.Time
	}

	return result
}
