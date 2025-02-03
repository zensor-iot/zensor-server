package domain

import "zensor-server/internal/infra/utils"

type Device struct {
	ID     string
	Name   string
	AppEUI string
	DevEUI string
	AppKey string
}

func NewDeviceBuilder() *deviceBuilder {
	return &deviceBuilder{}
}

type deviceBuilder struct {
	actions []deviceHandler
}

type deviceHandler func(v *Device) error

func (b *deviceBuilder) WithName(value string) *deviceBuilder {
	b.actions = append(b.actions, func(d *Device) error {
		d.Name = value
		return nil
	})
	return b
}

func (b *deviceBuilder) Build() (Device, error) {
	result := Device{
		ID:     utils.GenerateUUID(),
		DevEUI: utils.GenerateHEX(8),
		AppEUI: utils.GenerateHEX(8),
		AppKey: utils.GenerateHEX(16),
	}
	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return Device{}, err
		}
	}
	return result, nil
}
