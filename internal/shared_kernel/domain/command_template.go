package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
)

// CommandTemplate is a Value Object that represents a command template
// It contains all the command structure except instance-specific fields
// and includes a WaitFor field to calculate DispatchAfter when creating tasks
type CommandTemplate struct {
	Device   Device
	Port     Port
	Priority CommandPriority
	Payload  CommandPayload
	WaitFor  time.Duration // Duration to wait before dispatching the command
}

// NewCommandTemplateBuilder creates a new command template builder
func NewCommandTemplateBuilder() *commandTemplateBuilder {
	return &commandTemplateBuilder{}
}

type commandTemplateBuilder struct {
	actions []commandTemplateHandler
}

type commandTemplateHandler func(v *CommandTemplate) error

func (b *commandTemplateBuilder) WithDevice(value Device) *commandTemplateBuilder {
	b.actions = append(b.actions, func(d *CommandTemplate) error {
		d.Device = value
		return nil
	})
	return b
}

func (b *commandTemplateBuilder) WithPort(value Port) *commandTemplateBuilder {
	b.actions = append(b.actions, func(d *CommandTemplate) error {
		d.Port = value
		return nil
	})
	return b
}

func (b *commandTemplateBuilder) WithPriority(value CommandPriority) *commandTemplateBuilder {
	b.actions = append(b.actions, func(d *CommandTemplate) error {
		d.Priority = value
		return nil
	})
	return b
}

func (b *commandTemplateBuilder) WithPayload(value CommandPayload) *commandTemplateBuilder {
	b.actions = append(b.actions, func(d *CommandTemplate) error {
		d.Payload = value
		return nil
	})
	return b
}

func (b *commandTemplateBuilder) WithWaitFor(value time.Duration) *commandTemplateBuilder {
	b.actions = append(b.actions, func(d *CommandTemplate) error {
		d.WaitFor = value
		return nil
	})
	return b
}

func (b *commandTemplateBuilder) Build() (CommandTemplate, error) {
	result := CommandTemplate{
		Port: _defaultPort,
	}
	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return CommandTemplate{}, err
		}
	}
	return result, nil
}

// ToCommand converts a CommandTemplate to a Command with calculated DispatchAfter
func (ct CommandTemplate) ToCommand(task Task, baseTime time.Time) Command {
	dispatchAfter := baseTime.Add(ct.WaitFor)

	return Command{
		ID:            ID(utils.GenerateUUID()),
		Version:       1,
		Device:        ct.Device,
		Task:          task,
		Port:          ct.Port,
		Priority:      ct.Priority,
		Payload:       ct.Payload,
		DispatchAfter: utils.Time{Time: dispatchAfter},
		Ready:         false,
		Sent:          false,
	}
}
