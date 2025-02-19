package dto

import (
	"time"

	"github.com/vmihailenco/msgpack/v5"
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
	RawPayload     []byte         `json:"frm_payload"`
	DecodedPayload map[string]any `json:"decoded_payload"`
}

var (
	codeToNameMapping = map[string]string{
		"t": "temperature",
		"h": "humidity",
	}
)

func (m *UplinkMessage) FromMessagePack() any {
	temp := make(map[string][]byte)
	msgpack.Unmarshal(m.RawPayload, &temp)
	m.DecodedPayload = make(map[string]any)
	for k, v := range temp {
		m.DecodedPayload[codeToNameMapping[k]] = float64(v[1]) + float64(v[2])/100
	}
	return m.DecodedPayload
}

// 129 161 116 196 3 1 28 248 132 247
