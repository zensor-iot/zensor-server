package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/robfig/cron/v3"
)

const (
	_defaultTimezone     = "UTC"
	_scheduledTasksTopic = "scheduled_tasks"
)

func NewScheduledTaskWorker(
	ticker *time.Ticker,
	scheduledTaskRepository ScheduledTaskRepository,
	taskService TaskService,
	deviceService DeviceService,
	tenantConfigurationService TenantConfigurationService,
	broker async.InternalBroker,
) *ScheduledTaskWorker {
	return &ScheduledTaskWorker{
		ticker:                     ticker,
		scheduledTaskRepository:    scheduledTaskRepository,
		taskService:                taskService,
		deviceService:              deviceService,
		tenantConfigurationService: tenantConfigurationService,
		broker:                     broker,
		cronParser:                 cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

var _ async.Worker = &ScheduledTaskWorker{}

type ScheduledTaskWorker struct {
	ticker                     *time.Ticker
	scheduledTaskRepository    ScheduledTaskRepository
	taskService                TaskService
	deviceService              DeviceService
	tenantConfigurationService TenantConfigurationService
	broker                     async.InternalBroker
	cronParser                 cron.Parser
}

func (w *ScheduledTaskWorker) Run(ctx context.Context, done func()) {
	slog.Debug("scheduled task worker started")
	defer done()
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			slog.Info("scheduled task worker cancelled")
			wg.Wait()
			return
		case <-w.ticker.C:
			wg.Add(1)
			tickCtx := context.Background()
			w.evaluateSchedules(tickCtx, wg.Done)
		}
	}
}

func (w *ScheduledTaskWorker) evaluateSchedules(ctx context.Context, done func()) {
	slog.Debug("evaluating scheduled tasks", slog.Time("time", time.Now()))
	defer done()

	scheduledTasks, err := w.scheduledTaskRepository.FindAllActive(ctx)
	if err != nil {
		slog.Error("finding active scheduled tasks", slog.Any("error", err))
		return
	}

	slog.Debug("found scheduled tasks", slog.Int("count", len(scheduledTasks)))
	for _, scheduledTask := range scheduledTasks {
		w.evaluateAndMaybeCreateTask(ctx, scheduledTask)
	}

	slog.Debug("scheduled task evaluation completed", slog.Time("time", time.Now()))
}

func (w *ScheduledTaskWorker) evaluateAndMaybeCreateTask(ctx context.Context, scheduledTask domain.ScheduledTask) {
	slog.Debug("evaluating scheduled task", slog.String("scheduled_task_id", scheduledTask.ID.String()))
	if !scheduledTask.IsActive {
		return
	}

	var lastExecuted time.Time
	if scheduledTask.LastExecutedAt != nil {
		lastExecuted = scheduledTask.LastExecutedAt.Time
	}
	if lastExecuted.IsZero() {
		lastExecuted = scheduledTask.CreatedAt.Time
	}

	shouldExecute, err := w.shouldExecuteSchedule(ctx, scheduledTask, lastExecuted)
	if err != nil {
		slog.Error("evaluating schedule",
			slog.String("scheduled_task_id", scheduledTask.ID.String()),
			slog.String("schedule", scheduledTask.Schedule),
			slog.Any("error", err))
		return
	}

	if shouldExecute {
		w.createTaskFromScheduledTask(ctx, scheduledTask)
	}
}

func (w *ScheduledTaskWorker) shouldExecuteSchedule(ctx context.Context, scheduledTask domain.ScheduledTask, lastExecuted time.Time) (bool, error) {
	tenantConfig, err := w.tenantConfigurationService.GetOrCreateTenantConfiguration(ctx, scheduledTask.Tenant, _defaultTimezone)
	if err != nil {
		slog.Error("getting tenant configuration for timezone",
			slog.String("tenant_id", scheduledTask.Tenant.ID.String()),
			slog.String("error", err.Error()))
		tenantConfig, _ = domain.NewTenantConfigurationBuilder().
			WithTenantID(scheduledTask.Tenant.ID).
			WithTimezone(_defaultTimezone).
			Build()
	}

	location, err := time.LoadLocation(tenantConfig.Timezone)
	if err != nil {
		slog.Error("loading timezone location",
			slog.String("timezone", tenantConfig.Timezone),
			slog.String("tenant_id", scheduledTask.Tenant.ID.String()),
			slog.String("error", err.Error()))
		location = time.UTC
	}

	now := time.Now().In(location)
	lastExecutedInTZ := lastExecuted.In(location)

	var nextRun time.Time
	var scheduleInfo string

	switch scheduledTask.Scheduling.Type {
	case domain.SchedulingTypeInterval:
		nextRun, err = scheduledTask.CalculateNextExecution(tenantConfig.Timezone)
		if err != nil {
			return false, fmt.Errorf("calculating next interval execution: %w", err)
		}
		scheduleInfo = fmt.Sprintf("interval: every %d days at %s",
			*scheduledTask.Scheduling.DayInterval,
			*scheduledTask.Scheduling.ExecutionTime)

	case domain.SchedulingTypeCron:
		if scheduledTask.Schedule == "" {
			return false, errors.New("cron schedule is required for cron scheduling type")
		}

		scheduleSpec, err := w.cronParser.Parse(scheduledTask.Schedule)
		if err != nil {
			return false, fmt.Errorf("parsing cron schedule: %w", err)
		}
		nextRun = scheduleSpec.Next(lastExecutedInTZ)
		scheduleInfo = fmt.Sprintf("cron: %s", scheduledTask.Schedule)

	default:
		if scheduledTask.Schedule == "" {
			return false, errors.New("no valid scheduling configuration found")
		}

		scheduleSpec, err := w.cronParser.Parse(scheduledTask.Schedule)
		if err != nil {
			return false, fmt.Errorf("parsing cron schedule: %w", err)
		}
		nextRun = scheduleSpec.Next(lastExecutedInTZ)
		scheduleInfo = fmt.Sprintf("legacy cron: %s", scheduledTask.Schedule)
	}

	slog.Debug("evaluating schedule with timezone",
		slog.String("schedule_info", scheduleInfo),
		slog.String("scheduling_type", string(scheduledTask.Scheduling.Type)),
		slog.String("timezone", tenantConfig.Timezone),
		slog.Time("now", now),
		slog.Time("last_executed", lastExecutedInTZ),
		slog.Time("next_run", nextRun),
		slog.Bool("should_execute", nextRun.Before(now) || nextRun.Equal(now)))

	return nextRun.Before(now) || nextRun.Equal(now), nil
}

func (w *ScheduledTaskWorker) createTaskFromScheduledTask(ctx context.Context, scheduledTask domain.ScheduledTask) {
	device, err := w.deviceService.GetDevice(ctx, scheduledTask.Device.ID)
	if err != nil {
		slog.Error("getting device for scheduled task",
			slog.String("scheduled_task_id", scheduledTask.ID.String()),
			slog.String("device_id", scheduledTask.Device.ID.String()),
			slog.Any("error", err))
		return
	}

	commands := make([]domain.Command, len(scheduledTask.CommandTemplates))
	now := time.Now()

	for i, template := range scheduledTask.CommandTemplates {
		commandTemplate := domain.CommandTemplate{
			Device:   device,
			Port:     template.Port,
			Priority: template.Priority,
			Payload:  template.Payload,
			WaitFor:  template.WaitFor,
		}
		commands[i] = commandTemplate.ToCommand(domain.Task{}, now)
	}

	taskBuilder := domain.NewTaskBuilder()
	task, err := taskBuilder.
		WithDevice(device).
		WithCommands(commands).
		WithScheduledTask(&scheduledTask).
		Build()
	if err != nil {
		slog.Error("building task for scheduled task",
			slog.String("scheduled_task_id", scheduledTask.ID.String()),
			slog.Any("error", err))
		return
	}

	for i := range task.Commands {
		task.Commands[i].Task = task
	}

	err = w.taskService.Create(ctx, task)
	if err != nil {
		slog.Error("creating task from scheduled task",
			slog.String("scheduled_task_id", scheduledTask.ID.String()),
			slog.String("task_id", task.ID.String()),
			slog.Any("error", err))
		return
	}

	// Publish scheduled task executed event for metrics
	brokerMsg := async.BrokerMessage{
		Event: "scheduled_task_executed",
		Value: scheduledTask,
	}
	if err := w.broker.Publish(ctx, async.BrokerTopicName(_scheduledTasksTopic), brokerMsg); err != nil {
		slog.Error("failed to publish scheduled task executed event", slog.Any("error", err))
	}

	currentTime := utils.Time{Time: time.Now()}
	updatedScheduledTask := scheduledTask
	updatedScheduledTask.LastExecutedAt = &currentTime
	updatedScheduledTask.UpdatedAt = currentTime

	err = w.scheduledTaskRepository.Update(ctx, updatedScheduledTask)
	if err != nil {
		slog.Error("updating scheduled task last executed time",
			slog.String("scheduled_task_id", scheduledTask.ID.String()),
			slog.Any("error", err))
	}

	slog.Info("created task from scheduled task",
		slog.String("scheduled_task_id", scheduledTask.ID.String()),
		slog.String("task_id", task.ID.String()),
		slog.String("device_name", device.Name))

	// Metrics are now handled by MetricPublisherWorker
}

func (w *ScheduledTaskWorker) Shutdown() {
	slog.Warn("scheduled task worker shutdown is not yet implemented")
}
