package dto

import (
	"encoding/json"
	"fmt"
	"strings"
)

type VictronValue struct {
	Value float64 `json:"value"`
	Text  string  `json:"text"`
}

type VictronTelemetry struct {
	PortalID    string
	ServiceType string
	Instance    int
	Path        string
	Value       VictronValue
	Topic       string
}

type VictronServiceType string

const (
	ServiceBattery      VictronServiceType = "Battery"
	ServiceSolarCharger VictronServiceType = "SolarCharger"
	ServicePvInverter   VictronServiceType = "PvInverter"
	ServiceAcLoad       VictronServiceType = "AcLoad"
	ServiceDcLoad       VictronServiceType = "DcLoad"
	ServiceVebus        VictronServiceType = "Vebus"
	ServiceGenerator    VictronServiceType = "Generator"
	ServiceAlternator   VictronServiceType = "Alternator"
	ServiceTemperature  VictronServiceType = "Temperature"
	ServiceTank         VictronServiceType = "Tank"
	ServiceFuelLevel    VictronServiceType = "FuelLevel"
	ServiceGps          VictronServiceType = "Gps"
)

type VictronSystemSnapshot struct {
	PortalID      string                       `json:"portal_id"`
	Batteries     []BatteryData                `json:"batteries,omitempty"`
	SolarChargers []SolarChargerData           `json:"solar_chargers,omitempty"`
	PvInverters   []PvInverterData             `json:"pv_inverters,omitempty"`
	AcLoads       []LoadData                   `json:"ac_loads,omitempty"`
	DcLoads       []LoadData                   `json:"dc_loads,omitempty"`
	Vebus         []VebusData                  `json:"vebus,omitempty"`
	Generators    []GeneratorData              `json:"generators,omitempty"`
	Alternators   []AlternatorData             `json:"alternators,omitempty"`
	Temperatures  []TemperatureData            `json:"temperatures,omitempty"`
	Tanks         []TankData                   `json:"tanks,omitempty"`
	Raw           map[string]VictronValue      `json:"raw,omitempty"`
}

type BatteryData struct {
	Instance    int     `json:"instance"`
	Voltage     float64 `json:"voltage,omitempty"`
	Current     float64 `json:"current,omitempty"`
	Power       float64 `json:"power,omitempty"`
	Soc         float64 `json:"soc,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type SolarChargerData struct {
	Instance int     `json:"instance"`
	Power    float64 `json:"power,omitempty"`
	Voltage  float64 `json:"voltage,omitempty"`
	Current  float64 `json:"current,omitempty"`
}

type PvInverterData struct {
	Instance   int     `json:"instance"`
	Power      float64 `json:"power,omitempty"`
	Energy     float64 `json:"energy,omitempty"`
	Voltage    float64 `json:"voltage,omitempty"`
	Current    float64 `json:"current,omitempty"`
}

type LoadData struct {
	Instance int     `json:"instance"`
	Power    float64 `json:"power,omitempty"`
	Voltage  float64 `json:"voltage,omitempty"`
	Current  float64 `json:"current,omitempty"`
}

type VebusData struct {
	Instance    int     `json:"instance"`
	State       float64 `json:"state,omitempty"`
	Power       float64 `json:"power,omitempty"`
	Voltage     float64 `json:"voltage,omitempty"`
	Current     float64 `json:"current,omitempty"`
	Soc         float64 `json:"soc,omitempty"`
}

type GeneratorData struct {
	Instance int     `json:"instance"`
	Power    float64 `json:"power,omitempty"`
	Voltage  float64 `json:"voltage,omitempty"`
	Current  float64 `json:"current,omitempty"`
}

type AlternatorData struct {
	Instance int     `json:"instance"`
	Power    float64 `json:"power,omitempty"`
	Current  float64 `json:"current,omitempty"`
}

type TemperatureData struct {
	Instance    int     `json:"instance"`
	Temperature float64 `json:"temperature"`
}

type TankData struct {
	Instance int     `json:"instance"`
	Level    float64 `json:"level,omitempty"`
	Type     string  `json:"type,omitempty"`
}

func ParseVictronTopic(topic string) (serviceType string, instance int, path string, ok bool) {
	parts := strings.Split(topic, "/")
	if len(parts) < 6 {
		return "", 0, "", false
	}
	if parts[0] != "N" {
		return "", 0, "", false
	}
	serviceType = parts[4]
	_, err := fmt.Sscanf(parts[5], "%d", &instance)
	if err != nil {
		return "", 0, "", false
	}
	if len(parts) > 6 {
		path = strings.Join(parts[6:], "/")
	}
	return serviceType, instance, path, true
}

func ParseTelemetry(topic string, payload []byte) (VictronTelemetry, error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 4 || parts[0] != "N" {
		return VictronTelemetry{}, fmt.Errorf("invalid victron topic: %s", topic)
	}

	var value VictronValue
	if err := json.Unmarshal(payload, &value); err != nil {
		return VictronTelemetry{}, fmt.Errorf("parsing victron value from %s: %w", topic, err)
	}

	serviceType, instance, path, ok := ParseVictronTopic(topic)
	if !ok {
		return VictronTelemetry{}, fmt.Errorf("invalid victron topic: %s", topic)
	}

	return VictronTelemetry{
		PortalID:    parts[1],
		ServiceType: serviceType,
		Instance:    instance,
		Path:        path,
		Value:       value,
		Topic:       topic,
	}, nil
}

func ToSnakeName(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
