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
	_defaultTimezone               = "UTC"
	_executionsTopic               = "maintenance_executions"
	_executionCreatedEvent         = "execution_created"
	_executionCreationFailed       = "execution_creation_failed"
	_executionReadyForNotification = "execution_ready_for_notification"
	_executionOverdue              = "execution_overdue"
	_nextExecutionsCount           = 3
)

func NewExecutionWorker(
	ticker *time.Ticker,
	activityRepository ActivityRepository,
	executionRepository ExecutionRepository,
	executionService ExecutionService,
	tenantService controlPlaneUsecases.TenantService,
	tenantConfigurationService controlPlaneUsecases.TenantConfigurationService,
	broker async.InternalBroker,
) *ExecutionWorker {
	return &ExecutionWorker{
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

var _ async.Worker = &ExecutionWorker{}

type ExecutionWorker struct {
	ticker                     *time.Ticker
	activityRepository         ActivityRepository
	executionRepository        ExecutionRepository
	executionService           ExecutionService
	tenantService              controlPlaneUsecases.TenantService
	tenantConfigurationService controlPlaneUsecases.TenantConfigurationService
	broker                     async.InternalBroker
	cronParser                 cron.Parser
}

func (w *ExecutionWorker) Run(ctx context.Context, done func()) {
	slog.Info("execution worker started")
	defer done()
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			slog.Info("execution worker cancelled")
			wg.Wait()
			return
		case <-w.ticker.C:
			wg.Add(1)
			tickCtx := context.Background()
			w.scheduleExecutions(tickCtx, wg.Done)
			wg.Add(1)
			w.checkAndNotifyExecutions(tickCtx, wg.Done)
		}
	}
}

func (w *ExecutionWorker) ScheduleExecutions(ctx context.Context) {
	done := func() {}
	w.scheduleExecutions(ctx, done)
}

func (w *ExecutionWorker) scheduleExecutions(ctx context.Context, done func()) {
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

func (w *ExecutionWorker) processActivity(ctx context.Context, activity maintenanceDomain.Activity) {
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

func (w *ExecutionWorker) executionExists(ctx context.Context, activityID shareddomain.ID, scheduledDate time.Time) (bool, error) {
	_, err := w.executionRepository.FindByActivityAndScheduledDate(ctx, activityID, scheduledDate)
	if err != nil {
		if errors.Is(err, ErrExecutionNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (w *ExecutionWorker) createExecution(ctx context.Context, activity maintenanceDomain.Activity, scheduledDate time.Time, fieldValues map[string]any) error {
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

func (w *ExecutionWorker) buildFieldValuesFromActivity(activity maintenanceDomain.Activity) map[string]any {
	fieldValues := make(map[string]any)
	for _, field := range activity.Fields {
		if field.DefaultValue != nil {
			fieldValues[string(field.Name)] = *field.DefaultValue
		}
	}
	return fieldValues
}

func (w *ExecutionWorker) publishSuccessEvent(ctx context.Context, activity maintenanceDomain.Activity, scheduledDate time.Time) {
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

func (w *ExecutionWorker) publishFailureEvent(ctx context.Context, activity maintenanceDomain.Activity, err error) {
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

func (w *ExecutionWorker) checkAndNotifyExecutions(ctx context.Context, done func()) {
	slog.Info("checking executions for notifications", slog.Time("time", time.Now()))
	defer done()

	now := time.Now()
	w.checkReadyForNotification(ctx, now)
	w.checkOverdueExecutions(ctx, now)

	slog.Debug("execution notification check completed", slog.Time("time", time.Now()))
}

func (w *ExecutionWorker) checkReadyForNotification(ctx context.Context, currentDate time.Time) {
	executionsWithActivities, err := w.executionRepository.FindPendingExecutionsReadyForNotification(ctx, currentDate)
	if err != nil {
		slog.Error("finding executions ready for notification", slog.Any("error", err))
		return
	}

	slog.Debug("found executions ready for notification", slog.Int("count", len(executionsWithActivities)))

	for _, execWithActivity := range executionsWithActivities {
		daysBefore := int(execWithActivity.Execution.ScheduledDate.Time.Sub(currentDate).Hours() / 24)
		w.publishReadyForNotificationEvent(ctx, execWithActivity.Execution, execWithActivity.Activity, daysBefore)
	}
}

func (w *ExecutionWorker) checkOverdueExecutions(ctx context.Context, currentDate time.Time) {
	executionsWithActivities, err := w.executionRepository.FindOverdueExecutions(ctx)
	if err != nil {
		slog.Error("finding overdue executions", slog.Any("error", err))
		return
	}

	slog.Debug("found overdue executions", slog.Int("count", len(executionsWithActivities)))

	for _, execWithActivity := range executionsWithActivities {
		currentOverdueDays := execWithActivity.Execution.CalculateOverdueDays(currentDate)
		storedOverdueDays := int(execWithActivity.Execution.OverdueDays)

		if currentOverdueDays > storedOverdueDays {
			execWithActivity.Execution.OverdueDays = maintenanceDomain.OverdueDays(currentOverdueDays)
			err := w.executionRepository.Update(ctx, execWithActivity.Execution)
			if err != nil {
				slog.Error("updating execution overdue days",
					slog.String("execution_id", execWithActivity.Execution.ID.String()),
					slog.Any("error", err))
				continue
			}

			w.publishOverdueEvent(ctx, execWithActivity.Execution, execWithActivity.Activity, currentOverdueDays)
		}
	}
}

func (w *ExecutionWorker) publishReadyForNotificationEvent(ctx context.Context, execution maintenanceDomain.Execution, activity maintenanceDomain.Activity, daysBefore int) {
	brokerMsg := async.BrokerMessage{
		Event: _executionReadyForNotification,
		Value: map[string]any{
			"execution_id":   execution.ID.String(),
			"activity_id":    activity.ID.String(),
			"tenant_id":      activity.TenantID.String(),
			"scheduled_date": execution.ScheduledDate.Time,
			"days_before":    daysBefore,
		},
	}
	if err := w.broker.Publish(ctx, async.BrokerTopicName(_executionsTopic), brokerMsg); err != nil {
		slog.Error("failed to publish execution ready for notification event",
			slog.String("execution_id", execution.ID.String()),
			slog.Any("error", err))
	} else {
		slog.Info("published execution ready for notification event",
			slog.String("execution_id", execution.ID.String()),
			slog.Int("days_before", daysBefore))
	}
}

func (w *ExecutionWorker) publishOverdueEvent(ctx context.Context, execution maintenanceDomain.Execution, activity maintenanceDomain.Activity, overdueDays int) {
	brokerMsg := async.BrokerMessage{
		Event: _executionOverdue,
		Value: map[string]any{
			"execution_id":   execution.ID.String(),
			"activity_id":    activity.ID.String(),
			"tenant_id":      activity.TenantID.String(),
			"scheduled_date": execution.ScheduledDate.Time,
			"overdue_days":   overdueDays,
		},
	}
	if err := w.broker.Publish(ctx, async.BrokerTopicName(_executionsTopic), brokerMsg); err != nil {
		slog.Error("failed to publish execution overdue event",
			slog.String("execution_id", execution.ID.String()),
			slog.Any("error", err))
	} else {
		slog.Info("published execution overdue event",
			slog.String("execution_id", execution.ID.String()),
			slog.Int("overdue_days", overdueDays))
	}
}

func (w *ExecutionWorker) Shutdown() {
	slog.Warn("execution worker shutdown is not yet implemented")
}
