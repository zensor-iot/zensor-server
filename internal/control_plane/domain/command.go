package domain

import (
	"zensor-server/internal/infra/utils"
)

type Port uint8
type CommandPriority string
type CommandValue uint8

type CommandSequence struct {
	Commands []Command
}

type Command struct {
	ID            ID
	Device        Device
	Port          Port
	Priority      CommandPriority
	Payload       CommandPayload
	DispatchAfter utils.Time
	Ready         bool
	Sent          bool
	SentAt        utils.Time
}

type CommandPayload struct {
	Index Index
	Value CommandValue
}

func NewCommandBuilder() *commandBuilder {
	return &commandBuilder{}
}

type commandBuilder struct {
	actions []commandHandler
}

type commandHandler func(v *Command) error

func (b *commandBuilder) WithDevice(value Device) *commandBuilder {
	b.actions = append(b.actions, func(d *Command) error {
		d.Device = value
		return nil
	})
	return b
}

func (b *commandBuilder) WithPort(value Port) *commandBuilder {
	b.actions = append(b.actions, func(d *Command) error {
		d.Port = value
		return nil
	})
	return b
}

func (b *commandBuilder) WithPriority(value CommandPriority) *commandBuilder {
	b.actions = append(b.actions, func(d *Command) error {
		d.Priority = value
		return nil
	})
	return b
}

func (b *commandBuilder) WithPayload(value CommandPayload) *commandBuilder {
	b.actions = append(b.actions, func(d *Command) error {
		d.Payload = value
		return nil
	})
	return b
}

func (b *commandBuilder) WithDispatchAfter(value utils.Time) *commandBuilder {
	b.actions = append(b.actions, func(d *Command) error {
		d.DispatchAfter = value
		return nil
	})
	return b
}

func (b *commandBuilder) Build() (Command, error) {
	result := Command{
		ID:    ID(utils.GenerateUUID()),
		Ready: false,
		Sent:  false,
	}
	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return Command{}, err
		}
	}
	return result, nil
}
