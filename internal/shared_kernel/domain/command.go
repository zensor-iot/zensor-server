package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
)

const (
	_defaultPort Port = 15
	// _overlapBufferTime defines the time window to consider commands as potentially overlapping
	// This accounts for execution time, network delays, and safety margin
	_overlapBufferTime = 30 * time.Second
)

type Port uint8
type CommandPriority string
type CommandValue uint8

type CommandSequence struct {
	Commands []Command
}

type Command struct {
	ID            ID
	Version       Version
	Device        Device
	Task          Task
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

// OverlapsWith checks if this command overlaps with another command.
// Commands overlap if they target the same index (sensor/actuator) and their execution times could conflict.
func (c Command) OverlapsWith(other Command) bool {
	// Commands overlap if they target the same index (sensor/actuator)
	// and their execution times could conflict
	if c.Payload.Index != other.Payload.Index {
		return false // Different sensors/actuators, no overlap
	}

	// Calculate the effective time windows for each command
	cStart := c.DispatchAfter.Time
	cEnd := cStart.Add(_overlapBufferTime)

	otherStart := other.DispatchAfter.Time
	otherEnd := otherStart.Add(_overlapBufferTime)

	// Check if time windows overlap
	return cStart.Before(otherEnd) && otherStart.Before(cEnd)
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
		ID:      ID(utils.GenerateUUID()),
		Version: 1,
		Ready:   false,
		Sent:    false,
		Port:    _defaultPort,
	}
	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return Command{}, err
		}
	}
	return result, nil
}
