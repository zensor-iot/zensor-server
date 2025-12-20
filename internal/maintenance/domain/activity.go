package domain

import (
	"time"
	"zensor-server/internal/infra/utils"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type Activity struct {
	ID                     shareddomain.ID
	Version                shareddomain.Version
	TenantID               shareddomain.ID
	Type                   ActivityType
	CustomTypeName         *CustomTypeName
	Name                   shareddomain.Name
	Description            shareddomain.Description
	Schedule               Schedule
	NotificationDaysBefore Days
	Fields                 []FieldDefinition
	IsActive               bool
	CreatedAt              utils.Time
	UpdatedAt              utils.Time
	DeletedAt              *utils.Time
}

func (ma *Activity) IsDeleted() bool {
	return ma.DeletedAt != nil
}

func (ma *Activity) SoftDelete() {
	now := utils.Time{Time: time.Now()}
	ma.DeletedAt = &now
	ma.IsActive = false
	ma.UpdatedAt = now
}

func (ma *Activity) Activate() {
	ma.IsActive = true
	ma.UpdatedAt = utils.Time{Time: time.Now()}
}

func (ma *Activity) Deactivate() {
	ma.IsActive = false
	ma.UpdatedAt = utils.Time{Time: time.Now()}
}

func NewActivityBuilder() *activityBuilder {
	return &activityBuilder{}
}

type activityBuilder struct {
	actions []activityHandler
}

type activityHandler func(v *Activity) error

func (b *activityBuilder) WithTenantID(value shareddomain.ID) *activityBuilder {
	b.actions = append(b.actions, func(d *Activity) error {
		d.TenantID = value
		return nil
	})
	return b
}

func (b *activityBuilder) WithType(value ActivityType) *activityBuilder {
	b.actions = append(b.actions, func(d *Activity) error {
		d.Type = value
		return nil
	})
	return b
}

func (b *activityBuilder) WithCustomTypeName(value string) *activityBuilder {
	b.actions = append(b.actions, func(d *Activity) error {
		customTypeName := CustomTypeName(value)
		d.CustomTypeName = &customTypeName
		return nil
	})
	return b
}

func (b *activityBuilder) WithName(value string) *activityBuilder {
	b.actions = append(b.actions, func(d *Activity) error {
		d.Name = shareddomain.Name(value)
		return nil
	})
	return b
}

func (b *activityBuilder) WithDescription(value string) *activityBuilder {
	b.actions = append(b.actions, func(d *Activity) error {
		d.Description = shareddomain.Description(value)
		return nil
	})
	return b
}

func (b *activityBuilder) WithSchedule(value string) *activityBuilder {
	b.actions = append(b.actions, func(d *Activity) error {
		d.Schedule = Schedule(value)
		return nil
	})
	return b
}

func (b *activityBuilder) WithNotificationDaysBefore(value []int) *activityBuilder {
	b.actions = append(b.actions, func(d *Activity) error {
		d.NotificationDaysBefore = Days(value)
		return nil
	})
	return b
}

func (b *activityBuilder) WithFields(value []FieldDefinition) *activityBuilder {
	b.actions = append(b.actions, func(d *Activity) error {
		d.Fields = value
		return nil
	})
	return b
}

func (b *activityBuilder) Build() (Activity, error) {
	now := utils.Time{Time: time.Now()}
	result := Activity{
		ID:                     shareddomain.ID(utils.GenerateUUID()),
		Version:                1,
		NotificationDaysBefore: Days(make([]int, 0)),
		Fields:                 make([]FieldDefinition, 0),
		IsActive:               true,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return Activity{}, err
		}
	}

	return result, nil
}
