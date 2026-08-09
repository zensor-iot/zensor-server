package dto

import (
	"encoding/json"
	"fmt"
	"strings"
)

type VictronValue struct {
	Value float64
	Text  string
}

func (v *VictronValue) UnmarshalJSON(data []byte) error {
	var raw struct {
		Value any    `json:"value"`
		Text  string `json:"text"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	v.Text = raw.Text
	switch value := raw.Value.(type) {
	case float64:
		v.Value = value
	case string:
		v.Text = value
	}
	return nil
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
	ServiceBattery      VictronServiceType = "battery"
	ServiceSolarCharger VictronServiceType = "solarcharger"
	ServicePvInverter   VictronServiceType = "pvinverter"
	ServiceAcLoad       VictronServiceType = "acload"
	ServiceDcLoad       VictronServiceType = "dcload"
	ServiceVebus        VictronServiceType = "vebus"
	ServiceGenerator    VictronServiceType = "generator"
	ServiceAlternator   VictronServiceType = "alternator"
	ServiceTemperature  VictronServiceType = "temperature"
	ServiceTank         VictronServiceType = "tank"
	ServiceFuelLevel    VictronServiceType = "fuellevel"
	ServiceGps          VictronServiceType = "gps"
	ServiceSystem       VictronServiceType = "system"
)

type VictronSystemSnapshot struct {
	PortalID      string                  `json:"portal_id"`
	Batteries     []BatteryData           `json:"batteries,omitempty"`
	SolarChargers []SolarChargerData      `json:"solar_chargers,omitempty"`
	PvInverters   []PvInverterData        `json:"pv_inverters,omitempty"`
	AcLoads       []LoadData              `json:"ac_loads,omitempty"`
	DcLoads       []LoadData              `json:"dc_loads,omitempty"`
	Vebus         []VebusData             `json:"vebus,omitempty"`
	Generators    []GeneratorData         `json:"generators,omitempty"`
	Alternators   []AlternatorData        `json:"alternators,omitempty"`
	Temperatures  []TemperatureData       `json:"temperatures,omitempty"`
	Tanks         []TankData              `json:"tanks,omitempty"`
	System        SystemData              `json:"system"`
	Raw           map[string]VictronValue `json:"raw,omitempty"`
}

// SystemData holds the GX device's own pre-aggregated totals, published
// under the "system" service (e.g. N/<portalID>/system/0/Ac/Grid/L1/Power).
// These combine all connected devices and are what the Venus OS dashboard
// itself shows for grid, consumption, and overall battery/PV figures.
type SystemData struct {
	GridL1Power        float64 `json:"grid_l1_power,omitempty"`
	GridL2Power        float64 `json:"grid_l2_power,omitempty"`
	GridL3Power        float64 `json:"grid_l3_power,omitempty"`
	ConsumptionL1Power float64 `json:"consumption_l1_power,omitempty"`
	ConsumptionL2Power float64 `json:"consumption_l2_power,omitempty"`
	ConsumptionL3Power float64 `json:"consumption_l3_power,omitempty"`
	BatterySoc         float64 `json:"battery_soc,omitempty"`
	BatteryPower       float64 `json:"battery_power,omitempty"`
	PvPower            float64 `json:"pv_power,omitempty"`

	// AcActiveInSource mirrors Ac/ActiveIn/Source, where 240 means no input is
	// connected. It is nil until the GX reports it.
	AcActiveInSource *float64 `json:"ac_active_in_source,omitempty"`
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
	Instance int     `json:"instance"`
	Power    float64 `json:"power,omitempty"`
	Energy   float64 `json:"energy,omitempty"`
	Voltage  float64 `json:"voltage,omitempty"`
	Current  float64 `json:"current,omitempty"`
}

type LoadData struct {
	Instance int     `json:"instance"`
	Power    float64 `json:"power,omitempty"`
	Voltage  float64 `json:"voltage,omitempty"`
	Current  float64 `json:"current,omitempty"`
}

type VebusData struct {
	Instance int     `json:"instance"`
	State    float64 `json:"state,omitempty"`
	Power    float64 `json:"power,omitempty"`
	Voltage  float64 `json:"voltage,omitempty"`
	Current  float64 `json:"current,omitempty"`
	Soc      float64 `json:"soc,omitempty"`

	// ActiveIn holds the shore-power/grid input feeding the inverter.
	// ActiveInConnected and ActiveInput are nil until the GX publishes
	// Ac/ActiveIn/Connected and Ac/ActiveIn/ActiveInput respectively, which
	// distinguishes "reported as disconnected" from "never reported".
	// ActiveInput carries the input index, where 240 means none is active.
	ActiveInVoltage   float64  `json:"active_in_voltage,omitempty"`
	ActiveInCurrent   float64  `json:"active_in_current,omitempty"`
	ActiveInConnected *bool    `json:"active_in_connected,omitempty"`
	ActiveInput       *float64 `json:"active_in_input,omitempty"`
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
	if len(parts) < 4 {
		return "", 0, "", false
	}
	if parts[0] != "N" {
		return "", 0, "", false
	}
	serviceType = parts[2]
	_, err := fmt.Sscanf(parts[3], "%d", &instance)
	if err != nil {
		return "", 0, "", false
	}
	if len(parts) > 4 {
		path = strings.Join(parts[4:], "/")
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
