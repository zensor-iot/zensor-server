package domain

type Command struct {
	Device     Device
	RawPayload string
	Port       uint8
	Priority   string
}
