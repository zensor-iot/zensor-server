package dto

import (
	"slices"
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
	Port           uint8                   `json:"port"`
	RawPayload     []byte                  `json:"frm_payload"`
	DecodedPayload map[string][]SensorData `json:"decoded_payload"`
}

type SensorData struct {
	Index uint
	Value float64
}

var (
	codeToNameMapping = map[string]string{
		"t": "temperature",
		"h": "humidity",
		"w": "waterFlow",
	}
)

func (m *UplinkMessage) FromMessagePack() any {
	temp := make(map[string][]byte)
	msgpack.Unmarshal(m.RawPayload, &temp)
	m.DecodedPayload = make(map[string][]SensorData)
	for k, v := range temp {
		chunks := slices.Chunk(v, 3)
		m.DecodedPayload[codeToNameMapping[k]] = make([]SensorData, 0)
		for chunk := range chunks {
			if len(chunk) < 3 {
				break
			}
			m.DecodedPayload[codeToNameMapping[k]] = append(m.DecodedPayload[codeToNameMapping[k]], SensorData{
				Index: uint(chunk[0]),
				Value: float64(chunk[1]) + float64(chunk[2])/100,
			})
		}
	}
	return m.DecodedPayload
}
