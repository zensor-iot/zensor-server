package domain

type Port uint8
type CommandPriority string
type CommandValue uint8

type Command struct {
	Device   Device
	Port     Port
	Priority CommandPriority
	Payload  CommandPayload
}

type CommandPayload struct {
	Index Index
	Value CommandValue
}
