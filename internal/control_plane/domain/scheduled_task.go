package domain

import (
	"errors"
	"time"
	"zensor-server/internal/infra/utils"
)

type ScheduledTask struct {
	ID               ID
	Version          Version
	Tenant           Tenant
	Device           Device
	CommandTemplates []CommandTemplate
	Schedule         string // Cron format schedule
	IsActive         bool
	CreatedAt        utils.Time
	UpdatedAt        utils.Time
	LastExecutedAt   *utils.Time // When the scheduled task was last executed
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

	if result.Schedule == "" {
		return ScheduledTask{}, errors.New("schedule is required")
	}

	return result, nil
}
