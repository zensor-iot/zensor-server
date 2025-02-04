package dto

import (
	"time"
)

type Envelop struct {
	EndDeviceIDs   EndDeviceIDs  `json:"end_device_ids"`
	ReceivedAt     time.Time     `json:"received_at"`
	UplinkMessage  UplinkMessage `json:"uplink_message"`
	CorrelationIDs []string      `json:"correlation_ids"`
}

type EndDeviceIDs struct {
	DeviceID       string            `json:"device_id"`
	DevEUI         string            `json:"dev_eui"`
	JoinEUI        string            `json:"join_eui"`
	DevAddr        string            `json:"dev_addr"`
	ApplicationIDs map[string]string `json:"application_ids"`
}

type UplinkMessage struct {
	Port           uint8          `json:"port"`
	RawPayload     string         `json:"frm_payload"`
	DecodedPayload map[string]any `json:"decoded_payload"`
}
