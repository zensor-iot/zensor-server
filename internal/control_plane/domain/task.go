package domain

import (
	"errors"
	"zensor-server/internal/infra/utils"
)

type Task struct {
	ID       ID
	Version  Version
	Device   Device
	Commands []Command
}

func NewTaskBuilder() *taskBuilder {
	return &taskBuilder{}
}

type taskBuilder struct {
	actions []taskHandler
}

type taskHandler func(v *Task) error

func (b *taskBuilder) WithDevice(value Device) *taskBuilder {
	b.actions = append(b.actions, func(d *Task) error {
		d.Device = value
		return nil
	})
	return b
}

func (b *taskBuilder) WithCommands(value []Command) *taskBuilder {
	b.actions = append(b.actions, func(d *Task) error {
		d.Commands = value
		return nil
	})
	return b
}

func (b *taskBuilder) Build() (Task, error) {
	result := Task{
		ID:       ID(utils.GenerateUUID()),
		Version:  1,
		Commands: make([]Command, 0),
	}

	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return Task{}, err
		}
	}

	if result.Device.ID == "" {
		return Task{}, errors.New("device is required")
	}

	if len(result.Commands) == 0 {
		return Task{}, errors.New("commands are required")
	}

	return result, nil
}
