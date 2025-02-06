package dto

type Command struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	RawPayload string `json:"raw_payload"`
	Port       uint8  `json:"port"`
	Priority   string `json:"priority"`
}
