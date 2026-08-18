package domain

import (
	"errors"
	"fmt"
	"time"
	"zensor-server/internal/infra/utils"
)

var (
	errCalculateNextExecutionIntervalOnly         = errors.New("calculateNextExecution only supports interval scheduling")
	errTenantRequired                             = errors.New("tenant is required")
	errDeviceRequired                             = errors.New("device is required")
	errCommandTemplatesRequired                   = errors.New("command templates are required")
	errScheduleOrSchedulingConfigRequired         = errors.New("either schedule or scheduling configuration is required")
	errInitialDayRequiredForIntervalScheduling    = errors.New("initial_day is required for interval scheduling")
	errDayIntervalMustBeGreaterThanZero           = errors.New("day_interval must be greater than 0 for interval scheduling")
	errExecutionTimeRequiredForIntervalScheduling = errors.New("execution_time is required for interval scheduling")
	errInitialDayWithExecutionTimeMustBeInFuture  = errors.New("initial_day with execution_time must be in the future")
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
	st.UpdatedAt = now
}

func (st *ScheduledTask) CalculateNextExecution(tenantTimezone string) (time.Time, error) {
	if st.Scheduling.Type != SchedulingTypeInterval {
		return time.Time{}, errCalculateNextExecutionIntervalOnly
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
			st.Scheduling.InitialDay.In(location),
			location,
			true,
		), nil
	} else {
		return calculateNextIntervalExecution(
			st.Scheduling.InitialDay.Time,
			*st.Scheduling.DayInterval,
			hour,
			minute,
			st.LastExecutedAt.In(location),
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

	year, month, day := referenceTime.Date()
	lastExecutedDate := time.Date(year, month, day, 0, 0, 0, 0, location)
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
		return ScheduledTask{}, errTenantRequired
	}

	if result.Device.ID == "" {
		return ScheduledTask{}, errDeviceRequired
	}

	if len(result.CommandTemplates) == 0 {
		return ScheduledTask{}, errCommandTemplatesRequired
	}

	if result.Schedule == "" && result.Scheduling.Type == "" {
		return ScheduledTask{}, errScheduleOrSchedulingConfigRequired
	}

	if result.Schedule != "" && result.Scheduling.Type == "" {
		result.Scheduling = SchedulingConfiguration{
			Type: SchedulingTypeCron,
		}
	}

	if result.Scheduling.Type == SchedulingTypeInterval {
		if result.Scheduling.InitialDay == nil {
			return ScheduledTask{}, errInitialDayRequiredForIntervalScheduling
		}
		if result.Scheduling.DayInterval == nil || *result.Scheduling.DayInterval <= 0 {
			return ScheduledTask{}, errDayIntervalMustBeGreaterThanZero
		}
		if result.Scheduling.ExecutionTime == nil || *result.Scheduling.ExecutionTime == "" {
			return ScheduledTask{}, errExecutionTimeRequiredForIntervalScheduling
		}

		hour, minute, err := utils.ParseExecutionTime(*result.Scheduling.ExecutionTime)
		if err != nil {
			return ScheduledTask{}, fmt.Errorf("invalid execution_time format: %w", err)
		}

		firstExecutionTime := time.Date(
			result.Scheduling.InitialDay.Year(),
			result.Scheduling.InitialDay.Month(),
			result.Scheduling.InitialDay.Day(),
			hour,
			minute,
			0,
			0,
			result.Scheduling.InitialDay.Location(),
		)

		if firstExecutionTime.UTC().Before(time.Now().UTC()) {
			return ScheduledTask{}, errInitialDayWithExecutionTimeMustBeInFuture
		}
	}

	return result, nil
}
