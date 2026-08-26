package dto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Envelop struct {
	EndDeviceIDs   EndDeviceIDs  `json:"end_device_ids"`
	ReceivedAt     time.Time     `json:"received_at"`
	UplinkMessage  UplinkMessage `json:"uplink_message"`
	CorrelationIDs []string      `json:"correlation_ids"`
	Error          Error         `json:"error"`
}

type Error struct {
	Namespace     string `json:"namespace"`
	Name          string `json:"name"`
	MessageFormat string `json:"message_format"`
	CorrelationID string `json:"correlation_id"`
	Code          int    `json:"code"`
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
	DecodedPayload map[string][]SensorData `json:"decoded_payload,omitempty"`
}

type SensorData struct {
	Index uint    `json:"index"`
	Value float64 `json:"value"`
}

// UnmarshalJSON decodes an uplink message tolerating any decoded_payload shape
// produced by a device payload formatter. Entries that cannot be interpreted as
// sensor readings are skipped instead of failing the whole message.
func (m *UplinkMessage) UnmarshalJSON(data []byte) error {
	type uplinkMessageAlias struct {
		Port           uint8                      `json:"port"`
		RawPayload     []byte                     `json:"frm_payload"`
		DecodedPayload map[string]json.RawMessage `json:"decoded_payload,omitempty"`
	}

	var alias uplinkMessageAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return fmt.Errorf("unmarshalling uplink message: %w", err)
	}

	m.Port = alias.Port
	m.RawPayload = alias.RawPayload
	m.DecodedPayload = nil

	if alias.DecodedPayload == nil {
		return nil
	}

	m.DecodedPayload = make(map[string][]SensorData, len(alias.DecodedPayload))
	for key, raw := range alias.DecodedPayload {
		readings, ok := parseSensorDataList(raw)
		if !ok {
			slog.Warn("unsupported decoded payload entry", slog.String("key", key), slog.String("value", string(raw)))
			continue
		}
		m.DecodedPayload[key] = readings
	}

	return nil
}

func parseSensorDataList(raw json.RawMessage) ([]SensorData, bool) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, false
	}

	switch trimmed[0] {
	case 'n':
		return nil, true
	case '[':
		return parseSensorDataArray(trimmed)
	case '{':
		reading, ok := parseSensorDataObject(trimmed)
		if !ok {
			return nil, false
		}
		return []SensorData{reading}, true
	default:
		value, ok := parseSensorValue(trimmed)
		if !ok {
			return nil, false
		}
		return []SensorData{{Value: value}}, true
	}
}

func parseSensorDataArray(raw json.RawMessage) ([]SensorData, bool) {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, false
	}

	readings := make([]SensorData, 0, len(items))
	for index, item := range items {
		trimmed := bytes.TrimSpace(item)
		if len(trimmed) == 0 {
			return nil, false
		}

		if trimmed[0] == '{' {
			reading, ok := parseSensorDataObject(trimmed)
			if !ok {
				return nil, false
			}
			readings = append(readings, reading)
			continue
		}

		value, ok := parseSensorValue(trimmed)
		if !ok {
			return nil, false
		}
		readings = append(readings, SensorData{Index: uint(index), Value: value})
	}

	return readings, true
}

func parseSensorDataObject(raw json.RawMessage) (SensorData, bool) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return SensorData{}, false
	}

	if _, found := fields["value"]; !found {
		return SensorData{}, false
	}

	var reading SensorData
	if err := json.Unmarshal(raw, &reading); err != nil {
		return SensorData{}, false
	}

	return reading, true
}

func parseSensorValue(raw json.RawMessage) (float64, bool) {
	var number float64
	if err := json.Unmarshal(raw, &number); err == nil {
		return number, true
	}

	var flag bool
	if err := json.Unmarshal(raw, &flag); err == nil {
		if flag {
			return 1, true
		}
		return 0, true
	}

	return 0, false
}

var codeToNameMapping = map[string]string{
	"t": "temperature",
	"h": "humidity",
	"w": "waterFlow",
	"r": "relay",
}

func (m *UplinkMessage) FromMessagePack() any {
	temp := make(map[string][]byte)
	if err := msgpack.Unmarshal(m.RawPayload, &temp); err != nil {
		return nil
	}
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
