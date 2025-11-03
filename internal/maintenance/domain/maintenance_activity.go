package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
	sharedDomain "zensor-server/internal/shared_kernel/domain"
)

type MaintenanceActivity struct {
	ID                     sharedDomain.ID
	Version                sharedDomain.Version
	TenantID               sharedDomain.ID
	Type                   ActivityType
	CustomTypeName         *CustomTypeName
	Name                   sharedDomain.Name
	Description            sharedDomain.Description
	Schedule               Schedule
	NotificationDaysBefore Days
	Fields                 []FieldDefinition
	IsActive               bool
	CreatedAt              utils.Time
	UpdatedAt              utils.Time
	DeletedAt              *utils.Time
}

func (ma *MaintenanceActivity) IsDeleted() bool {
	return ma.DeletedAt != nil
}

func (ma *MaintenanceActivity) SoftDelete() {
	now := utils.Time{Time: time.Now()}
	ma.DeletedAt = &now
	ma.IsActive = false
	ma.UpdatedAt = now
}

func (ma *MaintenanceActivity) Activate() {
	ma.IsActive = true
	ma.UpdatedAt = utils.Time{Time: time.Now()}
}

func (ma *MaintenanceActivity) Deactivate() {
	ma.IsActive = false
	ma.UpdatedAt = utils.Time{Time: time.Now()}
}

func NewMaintenanceActivityBuilder() *maintenanceActivityBuilder {
	return &maintenanceActivityBuilder{}
}

type maintenanceActivityBuilder struct {
	actions []maintenanceActivityHandler
}

type maintenanceActivityHandler func(v *MaintenanceActivity) error

func (b *maintenanceActivityBuilder) WithTenantID(value sharedDomain.ID) *maintenanceActivityBuilder {
	b.actions = append(b.actions, func(d *MaintenanceActivity) error {
		d.TenantID = value
		return nil
	})
	return b
}

func (b *maintenanceActivityBuilder) WithType(value ActivityType) *maintenanceActivityBuilder {
	b.actions = append(b.actions, func(d *MaintenanceActivity) error {
		d.Type = value
		return nil
	})
	return b
}

func (b *maintenanceActivityBuilder) WithCustomTypeName(value string) *maintenanceActivityBuilder {
	b.actions = append(b.actions, func(d *MaintenanceActivity) error {
		customTypeName := CustomTypeName(value)
		d.CustomTypeName = &customTypeName
		return nil
	})
	return b
}

func (b *maintenanceActivityBuilder) WithName(value string) *maintenanceActivityBuilder {
	b.actions = append(b.actions, func(d *MaintenanceActivity) error {
		d.Name = sharedDomain.Name(value)
		return nil
	})
	return b
}

func (b *maintenanceActivityBuilder) WithDescription(value string) *maintenanceActivityBuilder {
	b.actions = append(b.actions, func(d *MaintenanceActivity) error {
		d.Description = sharedDomain.Description(value)
		return nil
	})
	return b
}

func (b *maintenanceActivityBuilder) WithSchedule(value string) *maintenanceActivityBuilder {
	b.actions = append(b.actions, func(d *MaintenanceActivity) error {
		d.Schedule = Schedule(value)
		return nil
	})
	return b
}

func (b *maintenanceActivityBuilder) WithNotificationDaysBefore(value []int) *maintenanceActivityBuilder {
	b.actions = append(b.actions, func(d *MaintenanceActivity) error {
		d.NotificationDaysBefore = Days(value)
		return nil
	})
	return b
}

func (b *maintenanceActivityBuilder) WithFields(value []FieldDefinition) *maintenanceActivityBuilder {
	b.actions = append(b.actions, func(d *MaintenanceActivity) error {
		d.Fields = value
		return nil
	})
	return b
}

func (b *maintenanceActivityBuilder) Build() (MaintenanceActivity, error) {
	now := utils.Time{Time: time.Now()}
	result := MaintenanceActivity{
		ID:                     sharedDomain.ID(utils.GenerateUUID()),
		Version:                1,
		NotificationDaysBefore: Days(make([]int, 0)),
		Fields:                 make([]FieldDefinition, 0),
		IsActive:               true,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return MaintenanceActivity{}, err
		}
	}

	return result, nil
}
