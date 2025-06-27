package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/async"
	"zensor-server/internal/infra/utils"

	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	_metricKeyScheduledTasks = "scheduled_tasks"
)

func NewScheduledTaskWorker(
	ticker *time.Ticker,
	scheduledTaskRepository ScheduledTaskRepository,
	taskService TaskService,
	deviceService DeviceService,
	broker async.InternalBroker,
) *ScheduledTaskWorker {
	return &ScheduledTaskWorker{
		ticker:                  ticker,
		scheduledTaskRepository: scheduledTaskRepository,
		taskService:             taskService,
		deviceService:           deviceService,
		broker:                  broker,
		metricCounters:          make(map[string]metric.Float64Counter),
		cronParser:              cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

var _ async.Worker = &ScheduledTaskWorker{}

type ScheduledTaskWorker struct {
	ticker                  *time.Ticker
	scheduledTaskRepository ScheduledTaskRepository
	taskService             TaskService
	deviceService           DeviceService
	broker                  async.InternalBroker
	metricCounters          map[string]metric.Float64Counter
	cronParser              cron.Parser
}

func (w *ScheduledTaskWorker) Run(ctx context.Context, done func()) {
	slog.Debug("scheduled task worker started")
	defer done()
	var wg sync.WaitGroup
	w.setupOtelCounters()

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

func (w *ScheduledTaskWorker) setupOtelCounters() {
	meter := otel.Meter("zensor_server")
	scheduledTaskCounter, _ := meter.Float64Counter(
		fmt.Sprintf("%s.%s", "zensor_server", "scheduled_tasks"),
		metric.WithDescription("zensor_server scheduled task counter"),
	)

	w.metricCounters[_metricKeyScheduledTasks] = scheduledTaskCounter
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
		slog.Debug("evaluating scheduled task", slog.String("scheduled_task_id", scheduledTask.ID.String()))
		if !scheduledTask.IsActive {
			continue
		}

		// Determine the last executed time for evaluation
		var lastExecuted time.Time
		if scheduledTask.LastExecutedAt != nil {
			lastExecuted = scheduledTask.LastExecutedAt.Time
		}
		if lastExecuted.IsZero() {
			lastExecuted = scheduledTask.CreatedAt.Time
		}

		shouldExecute, err := w.shouldExecuteSchedule(scheduledTask.Schedule, lastExecuted)
		if err != nil {
			slog.Error("evaluating schedule",
				slog.String("scheduled_task_id", scheduledTask.ID.String()),
				slog.String("schedule", scheduledTask.Schedule),
				slog.Any("error", err))
			continue
		}

		if shouldExecute {
			w.createTaskFromScheduledTask(ctx, scheduledTask)
		}
	}

	slog.Debug("scheduled task evaluation completed", slog.Time("time", time.Now()))
}

func (w *ScheduledTaskWorker) shouldExecuteSchedule(schedule string, lastExecuted time.Time) (bool, error) {
	scheduleSpec, err := w.cronParser.Parse(schedule)
	if err != nil {
		return false, fmt.Errorf("parsing cron schedule: %w", err)
	}

	now := time.Now()

	nextRun := scheduleSpec.Next(lastExecuted)

	return nextRun.Before(now) || nextRun.Equal(now), nil
}

func (w *ScheduledTaskWorker) createTaskFromScheduledTask(ctx context.Context, scheduledTask domain.ScheduledTask) {
	// Get the current device to ensure it exists and is up to date
	device, err := w.deviceService.GetDevice(ctx, scheduledTask.Device.ID)
	if err != nil {
		slog.Error("getting device for scheduled task",
			slog.String("scheduled_task_id", scheduledTask.ID.String()),
			slog.String("device_id", scheduledTask.Device.ID.String()),
			slog.Any("error", err))
		return
	}

	// Create a new task based on the scheduled task's command templates
	// We need to create new commands from templates with fresh IDs and calculated timestamps
	commands := make([]domain.Command, len(scheduledTask.CommandTemplates))
	now := time.Now()

	for i, template := range scheduledTask.CommandTemplates {
		// Update the device in the template to ensure it's current
		template.Device = device
		commands[i] = template.ToCommand(domain.Task{}, now)
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

	// Set the Task field on each command so they have the correct task reference
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

	// Update the last executed time for the scheduled task
	currentTime := utils.Time{Time: time.Now()}
	updatedScheduledTask := scheduledTask
	updatedScheduledTask.LastExecutedAt = &currentTime
	updatedScheduledTask.UpdatedAt = currentTime

	err = w.scheduledTaskRepository.Update(ctx, updatedScheduledTask)
	if err != nil {
		slog.Error("updating scheduled task last executed time",
			slog.String("scheduled_task_id", scheduledTask.ID.String()),
			slog.Any("error", err))
		// Don't return here as the task was already created successfully
	}

	slog.Info("created task from scheduled task",
		slog.String("scheduled_task_id", scheduledTask.ID.String()),
		slog.String("task_id", task.ID.String()),
		slog.String("device_name", device.Name))

	attributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String("zensor_server"),
		attribute.String("device_name", device.Name),
		attribute.String("scheduled_task_id", scheduledTask.ID.String()),
	}
	w.metricCounters[_metricKeyScheduledTasks].Add(ctx, 1, metric.WithAttributes(attributes...))
}

func (w *ScheduledTaskWorker) Shutdown() {
	slog.Warn("scheduled task worker shutdown is not yet implemented")
}
