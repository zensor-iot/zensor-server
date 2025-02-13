package domain

type SensorKind string

const (
	SensorKindTemperature SensorKind = "temperature"
)

type Sensor struct {
	Kind  SensorKind
	Index Index
}
