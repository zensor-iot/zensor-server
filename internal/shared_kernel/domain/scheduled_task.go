package domain

import (
	"errors"
	"fmt"
	"time"
	"zensor-server/internal/infra/utils"
)

type ScheduledTask struct {
	ID               ID
	Version          Version
	Tenant           Tenant
	Device           Device
	CommandTemplates []CommandTemplate
	Schedule         string                  // Cron format schedule (deprecated, use Scheduling instead)
	Scheduling       SchedulingConfiguration // New scheduling configuration
	IsActive         bool
	CreatedAt        utils.Time
	UpdatedAt        utils.Time
	LastExecutedAt   *utils.Time // When the scheduled task was last executed
	DeletedAt        *utils.Time // For soft deletion
}

// SchedulingConfiguration represents how a scheduled task should be executed
type SchedulingConfiguration struct {
	Type          SchedulingType // "cron" or "interval"
	InitialDay    *utils.Time    // Starting day for interval scheduling
	DayInterval   *int           // Days between executions (for interval scheduling)
	ExecutionTime *string        // Time of day (e.g., "02:00", "14:30")
}

// SchedulingType defines the type of scheduling
type SchedulingType string

const (
	SchedulingTypeCron     SchedulingType = "cron"
	SchedulingTypeInterval SchedulingType = "interval"
)

func (st *ScheduledTask) IsDeleted() bool {
	return st.DeletedAt != nil
}

func (st *ScheduledTask) SoftDelete() {
	now := utils.Time{Time: time.Now()}
	st.DeletedAt = &now
	st.IsActive = false
	st.Version++
	st.UpdatedAt = now
}

// CalculateNextExecution calculates the next execution time for interval-based scheduling
func (st *ScheduledTask) CalculateNextExecution(tenantTimezone string) (time.Time, error) {
	if st.Scheduling.Type != SchedulingTypeInterval {
		return time.Time{}, errors.New("calculateNextExecution only supports interval scheduling")
	}

	// Parse execution time (e.g., "02:00", "14:30")
	executionTime := *st.Scheduling.ExecutionTime
	location, err := time.LoadLocation(tenantTimezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("loading timezone %s: %w", tenantTimezone, err)
	}

	// Parse the execution time
	hour, minute, err := utils.ParseExecutionTime(executionTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing execution time %s: %w", executionTime, err)
	}

	// Determine reference time
	var referenceTime time.Time
	if st.LastExecutedAt != nil {
		referenceTime = st.LastExecutedAt.Time
	} else {
		referenceTime = st.CreatedAt.Time
	}

	// Convert to tenant timezone
	referenceTimeInTZ := referenceTime.In(location)

	// Calculate next execution
	return calculateNextIntervalExecution(
		st.Scheduling.InitialDay.Time,
		*st.Scheduling.DayInterval,
		hour,
		minute,
		referenceTimeInTZ,
		location,
	), nil
}

// calculateNextIntervalExecution calculates the next execution time for interval-based scheduling
func calculateNextIntervalExecution(
	initialDay time.Time,
	dayInterval int,
	hour, minute int,
	referenceTime time.Time,
	location *time.Location,
) time.Time {
	// If this is the first execution, find the next occurrence from initial day
	if referenceTime.Equal(initialDay) {
		// Find the next occurrence after the initial day
		candidate := time.Date(
			initialDay.Year(),
			initialDay.Month(),
			initialDay.Day(),
			hour,
			minute,
			0,
			0,
			location,
		)

		// If the candidate time is in the past or same as reference, move to next interval
		if candidate.Before(referenceTime) || candidate.Equal(referenceTime) {
			candidate = candidate.AddDate(0, 0, dayInterval)
		}

		return candidate
	}

	// For subsequent executions, calculate from last executed time
	lastExecutedDate := referenceTime.Truncate(24 * time.Hour)
	daysSinceInitial := int(lastExecutedDate.Sub(initialDay.Truncate(24*time.Hour)).Hours() / 24)

	// Calculate how many intervals have passed
	intervalsPassed := daysSinceInitial / dayInterval

	// Calculate the next interval
	nextInterval := intervalsPassed + 1
	nextExecutionDays := nextInterval * dayInterval

	// Calculate the next execution date
	nextExecutionDate := initialDay.AddDate(0, 0, nextExecutionDays)

	// Set the execution time
	nextExecution := time.Date(
		nextExecutionDate.Year(),
		nextExecutionDate.Month(),
		nextExecutionDate.Day(),
		hour,
		minute,
		0,
		0,
		location,
	)

	return nextExecution
}

func NewScheduledTaskBuilder() *scheduledTaskBuilder {
	return &scheduledTaskBuilder{}
}

type scheduledTaskBuilder struct {
	actions []scheduledTaskHandler
}

type scheduledTaskHandler func(v *ScheduledTask) error

func (b *scheduledTaskBuilder) WithTenant(value Tenant) *scheduledTaskBuilder {
	b.actions = append(b.actions, func(d *ScheduledTask) error {
		d.Tenant = value
		return nil
	})
	return b
}

func (b *scheduledTaskBuilder) WithDevice(value Device) *scheduledTaskBuilder {
	b.actions = append(b.actions, func(d *ScheduledTask) error {
		d.Device = value
		return nil
	})
	return b
}

func (b *scheduledTaskBuilder) WithCommandTemplates(value []CommandTemplate) *scheduledTaskBuilder {
	b.actions = append(b.actions, func(d *ScheduledTask) error {
		d.CommandTemplates = value
		return nil
	})
	return b
}

func (b *scheduledTaskBuilder) WithSchedule(value string) *scheduledTaskBuilder {
	b.actions = append(b.actions, func(d *ScheduledTask) error {
		d.Schedule = value
		return nil
	})
	return b
}

func (b *scheduledTaskBuilder) WithIsActive(value bool) *scheduledTaskBuilder {
	b.actions = append(b.actions, func(d *ScheduledTask) error {
		d.IsActive = value
		return nil
	})
	return b
}

func (b *scheduledTaskBuilder) WithLastExecutedAt(value *utils.Time) *scheduledTaskBuilder {
	b.actions = append(b.actions, func(d *ScheduledTask) error {
		d.LastExecutedAt = value
		return nil
	})
	return b
}

func (b *scheduledTaskBuilder) WithScheduling(value SchedulingConfiguration) *scheduledTaskBuilder {
	b.actions = append(b.actions, func(d *ScheduledTask) error {
		d.Scheduling = value
		return nil
	})
	return b
}

func (b *scheduledTaskBuilder) Build() (ScheduledTask, error) {
	now := utils.Time{Time: time.Now()}
	result := ScheduledTask{
		ID:               ID(utils.GenerateUUID()),
		Version:          1,
		IsActive:         true,
		CommandTemplates: make([]CommandTemplate, 0),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return ScheduledTask{}, err
		}
	}

	if result.Tenant.ID == "" {
		return ScheduledTask{}, errors.New("tenant is required")
	}

	if result.Device.ID == "" {
		return ScheduledTask{}, errors.New("device is required")
	}

	if len(result.CommandTemplates) == 0 {
		return ScheduledTask{}, errors.New("command templates are required")
	}

	// Validate scheduling configuration
	if result.Schedule == "" && result.Scheduling.Type == "" {
		return ScheduledTask{}, errors.New("either schedule or scheduling configuration is required")
	}

	// If using old cron format, convert to new scheduling configuration
	if result.Schedule != "" && result.Scheduling.Type == "" {
		result.Scheduling = SchedulingConfiguration{
			Type: SchedulingTypeCron,
		}
		// Keep the old Schedule field for backward compatibility
	}

	// Validate interval scheduling configuration
	if result.Scheduling.Type == SchedulingTypeInterval {
		if result.Scheduling.InitialDay == nil {
			return ScheduledTask{}, errors.New("initial_day is required for interval scheduling")
		}
		if result.Scheduling.DayInterval == nil || *result.Scheduling.DayInterval <= 0 {
			return ScheduledTask{}, errors.New("day_interval must be greater than 0 for interval scheduling")
		}
		if result.Scheduling.ExecutionTime == nil || *result.Scheduling.ExecutionTime == "" {
			return ScheduledTask{}, errors.New("execution_time is required for interval scheduling")
		}
	}

	return result, nil
}
