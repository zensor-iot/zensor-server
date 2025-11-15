package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	"zensor-server/internal/maintenance/persistence/internal"
	"zensor-server/internal/maintenance/usecases"
	"zensor-server/internal/shared_kernel/avro"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

const (
	_maintenanceActivitiesTopic = "maintenance_activities"
	_maintenanceExecutionsTopic = "maintenance_executions"
)

func NewMaintenanceActivityRepository(
	publisherFactory pubsub.PublisherFactory,
	orm sql.ORM,
) (*SimpleMaintenanceActivityRepository, error) {
	publisher, err := publisherFactory.New(_maintenanceActivitiesTopic, &avro.AvroMaintenanceActivity{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	err = orm.AutoMigrate(&internal.MaintenanceActivity{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleMaintenanceActivityRepository{
		publisher: publisher,
		orm:       orm,
	}, nil
}

var _ usecases.MaintenanceActivityRepository = (*SimpleMaintenanceActivityRepository)(nil)

type SimpleMaintenanceActivityRepository struct {
	publisher pubsub.Publisher
	orm       sql.ORM
}

func (r *SimpleMaintenanceActivityRepository) Create(ctx context.Context, activity maintenanceDomain.MaintenanceActivity) error {
	entity := internal.FromMaintenanceActivity(activity)

	err := r.orm.WithContext(ctx).Create(&entity).Error()
	if err != nil {
		return fmt.Errorf("creating maintenance activity in database: %w", err)
	}

	avroActivity := convertToAvroMaintenanceActivity(activity)

	slog.Debug("publishing maintenance activity to pubsub", slog.String("activity_id", activity.ID.String()))
	err = r.publisher.Publish(ctx, pubsub.Key(activity.ID), avroActivity)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}
	slog.Debug("published maintenance activity to pubsub", slog.String("activity_id", activity.ID.String()))

	return nil
}

func (r *SimpleMaintenanceActivityRepository) GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.MaintenanceActivity, error) {
	var entity internal.MaintenanceActivity
	err := r.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id.String()).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return maintenanceDomain.MaintenanceActivity{}, usecases.ErrMaintenanceActivityNotFound
	}

	if err != nil {
		return maintenanceDomain.MaintenanceActivity{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleMaintenanceActivityRepository) FindAllByTenant(
	ctx context.Context,
	tenantID shareddomain.ID,
	pagination usecases.Pagination,
) ([]maintenanceDomain.MaintenanceActivity, int, error) {
	var total int64
	query := r.orm.WithContext(ctx).Model(&internal.MaintenanceActivity{})

	err := query.Where("tenant_id = ? AND deleted_at IS NULL", tenantID.String()).Count(&total).Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.MaintenanceActivity
	err = query.
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID.String()).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&entities).
		Error()

	if err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	result := make([]maintenanceDomain.MaintenanceActivity, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, int(total), nil
}

func (r *SimpleMaintenanceActivityRepository) Update(ctx context.Context, activity maintenanceDomain.MaintenanceActivity) error {
	entity := internal.FromMaintenanceActivity(activity)

	err := r.orm.WithContext(ctx).Save(&entity).Error()
	if err != nil {
		return fmt.Errorf("updating maintenance activity in database: %w", err)
	}

	avroActivity := convertToAvroMaintenanceActivity(activity)

	err = r.publisher.Publish(ctx, pubsub.Key(activity.ID), avroActivity)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
	}

	return nil
}

func (r *SimpleMaintenanceActivityRepository) Delete(ctx context.Context, id shareddomain.ID) error {
	activity, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	activity.SoftDelete()
	return r.Update(ctx, activity)
}

func convertToAvroMaintenanceActivity(activity maintenanceDomain.MaintenanceActivity) *avro.AvroMaintenanceActivity {
	avroFields := make([]avro.AvroMaintenanceFieldDefinition, len(activity.Fields))
	for i, field := range activity.Fields {
		avroField := avro.AvroMaintenanceFieldDefinition{
			Name:        string(field.Name),
			DisplayName: string(field.DisplayName),
			Type:        string(field.Type),
			IsRequired:  field.IsRequired,
		}
		if field.DefaultValue != nil {
			if str, err := json.Marshal(field.DefaultValue); err == nil {
				defaultStr := string(str)
				avroField.DefaultValue = &defaultStr
			}
		}
		avroFields[i] = avroField
	}

	return &avro.AvroMaintenanceActivity{
		ID:                     activity.ID.String(),
		Version:                int(activity.Version),
		TenantID:               activity.TenantID.String(),
		TypeName:               string(activity.Type.Name),
		Name:                   string(activity.Name),
		Description:            string(activity.Description),
		Schedule:               string(activity.Schedule),
		NotificationDaysBefore: []int(activity.NotificationDaysBefore),
		Fields:                 avroFields,
		IsActive:               activity.IsActive,
		CreatedAt:              activity.CreatedAt.Time,
		UpdatedAt:              activity.UpdatedAt.Time,
	}
}
