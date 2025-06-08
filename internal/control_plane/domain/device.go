package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
)

type Device struct {
	ID                    ID
	Name                  string
	DisplayName           string // User-friendly name that can be edited in tenant portal
	AppEUI                string
	DevEUI                string
	AppKey                string
	TenantID              *ID // Optional tenant association, nil means orphan device
	Sector                *Sector
	EvaluationRules       []EvaluationRule
	LastMessageReceivedAt utils.Time
}

func (d *Device) AddEvaluationRule(evaluationRule EvaluationRule) {
	d.EvaluationRules = append(d.EvaluationRules, evaluationRule)
}

func (d *Device) AdoptToTenant(tenantID ID) {
	d.TenantID = &tenantID
}

func (d *Device) IsOrphan() bool {
	return d.TenantID == nil
}

func (d *Device) BelongsToTenant(tenantID ID) bool {
	return d.TenantID != nil && *d.TenantID == tenantID
}

func (d *Device) UpdateDisplayName(displayName string) {
	d.DisplayName = displayName
}

// UpdateLastMessageReceivedAt updates the timestamp when a message was last received from TTN
func (d *Device) UpdateLastMessageReceivedAt(timestamp utils.Time) {
	d.LastMessageReceivedAt = timestamp
}

// IsOnline returns true if the device received a message within the last 5 minutes
func (d *Device) IsOnline() bool {
	if d.LastMessageReceivedAt.IsZero() {
		return false
	}

	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	return d.LastMessageReceivedAt.After(fiveMinutesAgo)
}

// GetStatus returns "online" or "offline" based on last message timestamp
func (d *Device) GetStatus() string {
	if d.IsOnline() {
		return "online"
	}
	return "offline"
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

func (b *deviceBuilder) WithDisplayName(value string) *deviceBuilder {
	b.actions = append(b.actions, func(d *Device) error {
		d.DisplayName = value
		return nil
	})
	return b
}

func (b *deviceBuilder) WithTenant(tenantID ID) *deviceBuilder {
	b.actions = append(b.actions, func(d *Device) error {
		d.TenantID = &tenantID
		return nil
	})
	return b
}

func (b *deviceBuilder) Build() (Device, error) {
	result := Device{
		ID:              ID(utils.GenerateUUID()),
		DevEUI:          utils.GenerateHEX(8),
		AppEUI:          utils.GenerateHEX(8),
		AppKey:          utils.GenerateHEX(16),
		EvaluationRules: make([]EvaluationRule, 0),
	}
	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return Device{}, err
		}
	}
	return result, nil
}
