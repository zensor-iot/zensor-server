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
	Schedule         string
	Scheduling       SchedulingConfiguration
	IsActive         bool
	CreatedAt        utils.Time
	UpdatedAt        utils.Time
	LastExecutedAt   *utils.Time
	DeletedAt        *utils.Time
}

type SchedulingConfiguration struct {
	Type          SchedulingType
	InitialDay    *utils.Time
	DayInterval   *int
	ExecutionTime *string
}

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

func (st *ScheduledTask) CalculateNextExecution(tenantTimezone string) (time.Time, error) {
	if st.Scheduling.Type != SchedulingTypeInterval {
		return time.Time{}, errors.New("calculateNextExecution only supports interval scheduling")
	}

	executionTime := *st.Scheduling.ExecutionTime
	location, err := time.LoadLocation(tenantTimezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("loading timezone %s: %w", tenantTimezone, err)
	}

	hour, minute, err := utils.ParseExecutionTime(executionTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing execution time %s: %w", executionTime, err)
	}

	if st.LastExecutedAt == nil {
		return calculateNextIntervalExecution(
			st.Scheduling.InitialDay.Time,
			*st.Scheduling.DayInterval,
			hour,
			minute,
			st.Scheduling.InitialDay.Time.In(location),
			location,
			true,
		), nil
	} else {
		return calculateNextIntervalExecution(
			st.Scheduling.InitialDay.Time,
			*st.Scheduling.DayInterval,
			hour,
			minute,
			st.LastExecutedAt.Time.In(location),
			location,
			false,
		), nil
	}
}

func calculateNextIntervalExecution(
	initialDay time.Time,
	dayInterval int,
	hour, minute int,
	referenceTime time.Time,
	location *time.Location,
	isFirstExecution bool,
) time.Time {
	if isFirstExecution {
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

		if candidate.Before(referenceTime) {
			candidate = candidate.AddDate(0, 0, dayInterval)
		}

		return candidate
	}

	lastExecutedDate := referenceTime.Truncate(24 * time.Hour)
	nextExecutionDate := lastExecutedDate.AddDate(0, 0, dayInterval)

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

	if result.Schedule == "" && result.Scheduling.Type == "" {
		return ScheduledTask{}, errors.New("either schedule or scheduling configuration is required")
	}

	if result.Schedule != "" && result.Scheduling.Type == "" {
		result.Scheduling = SchedulingConfiguration{
			Type: SchedulingTypeCron,
		}
	}

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

		hour, minute, err := utils.ParseExecutionTime(*result.Scheduling.ExecutionTime)
		if err != nil {
			return ScheduledTask{}, fmt.Errorf("invalid execution_time format: %w", err)
		}

		firstExecutionTime := time.Date(
			result.Scheduling.InitialDay.Time.Year(),
			result.Scheduling.InitialDay.Time.Month(),
			result.Scheduling.InitialDay.Time.Day(),
			hour,
			minute,
			0,
			0,
			result.Scheduling.InitialDay.Time.Location(),
		)

		if firstExecutionTime.Before(time.Now()) {
			return ScheduledTask{}, errors.New("initial_day with execution_time must be in the future")
		}
	}

	return result, nil
}
