package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
	controlPlaneUsecases "zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/async"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"

	"github.com/robfig/cron/v3"
)

const (
	_defaultTimezone         = "UTC"
	_executionsTopic         = "maintenance_executions"
	_executionCreatedEvent   = "execution_created"
	_executionCreationFailed = "execution_creation_failed"
	_nextExecutionsCount     = 3
)

func NewExecutionScheduler(
	ticker *time.Ticker,
	activityRepository ActivityRepository,
	executionRepository ExecutionRepository,
	executionService ExecutionService,
	tenantService controlPlaneUsecases.TenantService,
	tenantConfigurationService controlPlaneUsecases.TenantConfigurationService,
	broker async.InternalBroker,
) *ExecutionScheduler {
	return &ExecutionScheduler{
		ticker:                     ticker,
		activityRepository:         activityRepository,
		executionRepository:        executionRepository,
		executionService:           executionService,
		tenantService:              tenantService,
		tenantConfigurationService: tenantConfigurationService,
		broker:                     broker,
		cronParser:                 cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

var _ async.Worker = &ExecutionScheduler{}

type ExecutionScheduler struct {
	ticker                     *time.Ticker
	activityRepository         ActivityRepository
	executionRepository        ExecutionRepository
	executionService           ExecutionService
	tenantService              controlPlaneUsecases.TenantService
	tenantConfigurationService controlPlaneUsecases.TenantConfigurationService
	broker                     async.InternalBroker
	cronParser                 cron.Parser
}

func (w *ExecutionScheduler) Run(ctx context.Context, done func()) {
	slog.Info("execution scheduler started")
	defer done()
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			slog.Info("execution scheduler cancelled")
			wg.Wait()
			return
		case <-w.ticker.C:
			wg.Add(1)
			tickCtx := context.Background()
			w.scheduleExecutions(tickCtx, wg.Done)
		}
	}
}

func (w *ExecutionScheduler) ScheduleExecutions(ctx context.Context) {
	done := func() {}
	w.scheduleExecutions(ctx, done)
}

func (w *ExecutionScheduler) scheduleExecutions(ctx context.Context, done func()) {
	slog.Info("scheduling executions", slog.Time("time", time.Now()))
	defer done()

	activities, err := w.activityRepository.FindAllActive(ctx)
	if err != nil {
		slog.Error("finding active activities", slog.Any("error", err))
		return
	}

	slog.Debug("found active activities", slog.Int("count", len(activities)))
	for _, activity := range activities {
		w.processActivity(ctx, activity)
	}

	slog.Debug("execution scheduling completed", slog.Time("time", time.Now()))
}

func (w *ExecutionScheduler) processActivity(ctx context.Context, activity maintenanceDomain.Activity) {
	slog.Debug("processing activity", slog.String("activity_id", activity.ID.String()))

	tenant, err := w.tenantService.GetTenant(ctx, activity.TenantID)
	if err != nil {
		slog.Error("getting tenant for activity",
			slog.String("activity_id", activity.ID.String()),
			slog.String("tenant_id", activity.TenantID.String()),
			slog.Any("error", err))
		w.publishFailureEvent(ctx, activity, fmt.Errorf("getting tenant: %w", err))
		return
	}

	tenantConfig, err := w.tenantConfigurationService.GetOrCreateTenantConfiguration(ctx, tenant, _defaultTimezone)
	if err != nil {
		slog.Error("getting tenant configuration for timezone",
			slog.String("tenant_id", tenant.ID.String()),
			slog.String("error", err.Error()))
		tenantConfig, _ = shareddomain.NewTenantConfigurationBuilder().
			WithTenantID(tenant.ID).
			WithTimezone(_defaultTimezone).
			Build()
	}

	location, err := time.LoadLocation(tenantConfig.Timezone)
	if err != nil {
		slog.Error("loading timezone location",
			slog.String("timezone", tenantConfig.Timezone),
			slog.String("tenant_id", tenant.ID.String()),
			slog.String("error", err.Error()))
		location = time.UTC
	}

	now := time.Now().In(location)

	scheduleSpec, err := w.cronParser.Parse(string(activity.Schedule))
	if err != nil {
		slog.Error("parsing cron schedule",
			slog.String("activity_id", activity.ID.String()),
			slog.String("schedule", string(activity.Schedule)),
			slog.Any("error", err))
		w.publishFailureEvent(ctx, activity, fmt.Errorf("parsing cron schedule: %w", err))
		return
	}

	fieldValues := w.buildFieldValuesFromActivity(activity)

	lastExecutionTime := now
	for range _nextExecutionsCount {
		nextTime := scheduleSpec.Next(lastExecutionTime)

		if nextTime.Before(now) || nextTime.Equal(now) {
			lastExecutionTime = nextTime
			continue
		}

		lastExecutionTime = nextTime

		exists, err := w.executionExists(ctx, activity.ID, nextTime)
		if err != nil {
			slog.Error("checking if execution exists",
				slog.String("activity_id", activity.ID.String()),
				slog.Time("scheduled_date", nextTime),
				slog.Any("error", err))
			w.publishFailureEvent(ctx, activity, fmt.Errorf("checking execution existence: %w", err))
			continue
		}

		if exists {
			slog.Debug("execution already exists, skipping",
				slog.String("activity_id", activity.ID.String()),
				slog.Time("scheduled_date", nextTime))
			continue
		}

		err = w.createExecution(ctx, activity, nextTime, fieldValues)
		if err != nil {
			slog.Error("creating execution",
				slog.String("activity_id", activity.ID.String()),
				slog.Time("scheduled_date", nextTime),
				slog.Any("error", err))
			w.publishFailureEvent(ctx, activity, fmt.Errorf("creating execution: %w", err))
			continue
		}

		slog.Info("execution created successfully",
			slog.String("activity_id", activity.ID.String()),
			slog.Time("scheduled_date", nextTime))
		w.publishSuccessEvent(ctx, activity, nextTime)
	}
}

func (w *ExecutionScheduler) executionExists(ctx context.Context, activityID shareddomain.ID, scheduledDate time.Time) (bool, error) {
	_, err := w.executionRepository.FindByActivityAndScheduledDate(ctx, activityID, scheduledDate)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (w *ExecutionScheduler) createExecution(ctx context.Context, activity maintenanceDomain.Activity, scheduledDate time.Time, fieldValues map[string]any) error {
	executionBuilder := maintenanceDomain.NewExecutionBuilder()
	execution, err := executionBuilder.
		WithActivityID(activity.ID).
		WithScheduledDate(scheduledDate).
		WithFieldValues(fieldValues).
		Build()
	if err != nil {
		return fmt.Errorf("building execution: %w", err)
	}

	return w.executionService.CreateExecution(ctx, execution)
}

func (w *ExecutionScheduler) buildFieldValuesFromActivity(activity maintenanceDomain.Activity) map[string]any {
	fieldValues := make(map[string]any)
	for _, field := range activity.Fields {
		if field.DefaultValue != nil {
			fieldValues[string(field.Name)] = *field.DefaultValue
		}
	}
	return fieldValues
}

func (w *ExecutionScheduler) publishSuccessEvent(ctx context.Context, activity maintenanceDomain.Activity, scheduledDate time.Time) {
	brokerMsg := async.BrokerMessage{
		Event: _executionCreatedEvent,
		Value: map[string]any{
			"activity_id":    activity.ID.String(),
			"scheduled_date": scheduledDate,
		},
	}
	if err := w.broker.Publish(ctx, async.BrokerTopicName(_executionsTopic), brokerMsg); err != nil {
		slog.Error("failed to publish execution created event", slog.Any("error", err))
	}
}

func (w *ExecutionScheduler) publishFailureEvent(ctx context.Context, activity maintenanceDomain.Activity, err error) {
	brokerMsg := async.BrokerMessage{
		Event: _executionCreationFailed,
		Value: map[string]any{
			"activity_id": activity.ID.String(),
			"error":       err.Error(),
		},
		Error: err,
	}
	if err := w.broker.Publish(ctx, async.BrokerTopicName(_executionsTopic), brokerMsg); err != nil {
		slog.Error("failed to publish execution creation failed event", slog.Any("error", err))
	}
}

func (w *ExecutionScheduler) Shutdown() {
	slog.Warn("execution scheduler shutdown is not yet implemented")
}
