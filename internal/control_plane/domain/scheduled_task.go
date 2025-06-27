package domain

import (
	"errors"
	"zensor-server/internal/infra/utils"
)

type ScheduledTask struct {
	ID       ID
	Version  Version
	Tenant   Tenant
	Device   Device
	Commands []Command
	Schedule string // Cron format schedule
	IsActive bool
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

func (b *scheduledTaskBuilder) WithCommands(value []Command) *scheduledTaskBuilder {
	b.actions = append(b.actions, func(d *ScheduledTask) error {
		d.Commands = value
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

func (b *scheduledTaskBuilder) Build() (ScheduledTask, error) {
	result := ScheduledTask{
		ID:       ID(utils.GenerateUUID()),
		Version:  1,
		IsActive: true,
		Commands: make([]Command, 0),
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

	if len(result.Commands) == 0 {
		return ScheduledTask{}, errors.New("commands are required")
	}

	if result.Schedule == "" {
		return ScheduledTask{}, errors.New("schedule is required")
	}

	return result, nil
}
