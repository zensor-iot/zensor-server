package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"zensor-server/internal/infra/sql"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	"zensor-server/internal/maintenance/persistence/internal"
	"zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

func NewExecutionRepository(orm sql.ORM) (*SimpleExecutionRepository, error) {
	err := orm.AutoMigrate(&internal.Execution{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &SimpleExecutionRepository{
		orm: orm,
	}, nil
}

var _ usecases.ExecutionRepository = (*SimpleExecutionRepository)(nil)

type SimpleExecutionRepository struct {
	orm sql.ORM
}

func (r *SimpleExecutionRepository) Create(ctx context.Context, execution maintenanceDomain.Execution) error {
	entity := internal.FromExecution(execution)

	err := r.orm.WithContext(ctx).Create(&entity).Error()
	if err != nil {
		return fmt.Errorf("creating maintenance execution in database: %w", err)
	}

	return nil
}

func (r *SimpleExecutionRepository) GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Execution, error) {
	var entity internal.Execution
	err := r.orm.
		WithContext(ctx).
		First(&entity, "id = ?", id.String()).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return maintenanceDomain.Execution{}, usecases.ErrExecutionNotFound
	}

	if err != nil {
		return maintenanceDomain.Execution{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleExecutionRepository) FindAllByActivity(
	ctx context.Context,
	activityID shareddomain.ID,
	pagination usecases.Pagination,
) ([]maintenanceDomain.Execution, int, error) {
	var total int64
	query := r.orm.WithContext(ctx).Model(&internal.Execution{})

	err := query.Where("activity_id = ? AND deleted_at IS NULL", activityID.String()).Count(&total).Error()
	if err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	var entities []internal.Execution
	err = query.
		Where("activity_id = ? AND deleted_at IS NULL", activityID.String()).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&entities).
		Error()

	if err != nil {
		return nil, 0, fmt.Errorf("database query: %w", err)
	}

	result := make([]maintenanceDomain.Execution, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, int(total), nil
}

func (r *SimpleExecutionRepository) FindByActivityAndScheduledDate(ctx context.Context, activityID shareddomain.ID, scheduledDate time.Time) (maintenanceDomain.Execution, error) {
	var entity internal.Execution
	err := r.orm.
		WithContext(ctx).
		Where("activity_id = ? AND scheduled_date = ? AND deleted_at IS NULL", activityID.String(), scheduledDate).
		First(&entity).
		Error()

	if errors.Is(err, sql.ErrRecordNotFound) {
		return maintenanceDomain.Execution{}, usecases.ErrExecutionNotFound
	}

	if err != nil {
		return maintenanceDomain.Execution{}, fmt.Errorf("database query: %w", err)
	}

	return entity.ToDomain(), nil
}

func (r *SimpleExecutionRepository) Update(ctx context.Context, execution maintenanceDomain.Execution) error {
	entity := internal.FromExecution(execution)

	err := r.orm.WithContext(ctx).Save(&entity).Error()
	if err != nil {
		return fmt.Errorf("updating maintenance execution in database: %w", err)
	}

	return nil
}

func (r *SimpleExecutionRepository) MarkCompleted(ctx context.Context, id shareddomain.ID, completedBy string) error {
	execution, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	execution.MarkCompleted(completedBy)
	return r.Update(ctx, execution)
}

func (r *SimpleExecutionRepository) FindAllOverdue(ctx context.Context, tenantID shareddomain.ID) ([]maintenanceDomain.Execution, error) {
	var entities []internal.Execution
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

	result := make([]maintenanceDomain.Execution, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, nil
}

func (r *SimpleExecutionRepository) FindAllDueSoon(ctx context.Context, tenantID shareddomain.ID, days int) ([]maintenanceDomain.Execution, error) {
	var entities []internal.Execution
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

	result := make([]maintenanceDomain.Execution, len(entities))
	for i, entity := range entities {
		result[i] = entity.ToDomain()
	}

	return result, nil
}

func (r *SimpleExecutionRepository) FindPendingExecutionsReadyForNotification(ctx context.Context, currentDate time.Time) ([]usecases.ExecutionWithActivity, error) {
	var executions []internal.Execution

	err := r.orm.
		WithContext(ctx).
		Model(&internal.Execution{}).
		Joins("JOIN maintenance_activities ON maintenance_executions.activity_id = maintenance_activities.id").
		Where("maintenance_executions.deleted_at IS NULL").
		Where("maintenance_executions.completed_at IS NULL").
		Where("maintenance_activities.deleted_at IS NULL").
		Where("maintenance_activities.is_active = ?", true).
		Where("maintenance_executions.scheduled_date > ?", currentDate).
		Find(&executions).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	if len(executions) == 0 {
		return nil, nil
	}

	activities, err := r.fetchActivitiesByExecutions(ctx, executions)
	if err != nil {
		return nil, err
	}

	result := make([]usecases.ExecutionWithActivity, 0, len(executions))
	for _, exec := range executions {
		activity, ok := activities[exec.ActivityID]
		if !ok {
			continue
		}

		execution := exec.ToDomain()
		domainActivity := activity.ToDomain()

		daysUntil := int(execution.ScheduledDate.Time.Sub(currentDate).Hours() / 24)
		notificationDays := []int(domainActivity.NotificationDaysBefore)

		for _, notificationDay := range notificationDays {
			if daysUntil == notificationDay {
				result = append(result, usecases.ExecutionWithActivity{
					Execution: execution,
					Activity:  domainActivity,
				})
				break
			}
		}
	}

	return result, nil
}

func (r *SimpleExecutionRepository) FindOverdueExecutions(ctx context.Context) ([]usecases.ExecutionWithActivity, error) {
	now := time.Now()

	var executions []internal.Execution
	err := r.orm.
		WithContext(ctx).
		Model(&internal.Execution{}).
		Joins("JOIN maintenance_activities ON maintenance_executions.activity_id = maintenance_activities.id").
		Where("maintenance_executions.deleted_at IS NULL").
		Where("maintenance_executions.completed_at IS NULL").
		Where("maintenance_activities.deleted_at IS NULL").
		Where("maintenance_activities.is_active = ?", true).
		Where("maintenance_executions.scheduled_date < ?", now).
		Find(&executions).
		Error()

	if err != nil {
		return nil, fmt.Errorf("database query: %w", err)
	}

	if len(executions) == 0 {
		return nil, nil
	}

	activities, err := r.fetchActivitiesByExecutions(ctx, executions)
	if err != nil {
		return nil, err
	}

	result := make([]usecases.ExecutionWithActivity, 0, len(executions))
	for _, exec := range executions {
		if activity, ok := activities[exec.ActivityID]; ok {
			result = append(result, usecases.ExecutionWithActivity{
				Execution: exec.ToDomain(),
				Activity:  activity.ToDomain(),
			})
		}
	}

	return result, nil
}

func (r *SimpleExecutionRepository) fetchActivitiesByExecutions(ctx context.Context, executions []internal.Execution) (map[string]internal.Activity, error) {
	activityIDs := make([]string, len(executions))
	for i, exec := range executions {
		activityIDs[i] = exec.ActivityID
	}

	var activities []internal.Activity
	err := r.orm.
		WithContext(ctx).
		Model(&internal.Activity{}).
		Where("id IN ?", activityIDs).
		Find(&activities).
		Error()
	if err != nil {
		return nil, fmt.Errorf("fetching activities: %w", err)
	}

	activityMap := make(map[string]internal.Activity, len(activities))
	for _, a := range activities {
		activityMap[a.ID] = a
	}

	return activityMap, nil
}
