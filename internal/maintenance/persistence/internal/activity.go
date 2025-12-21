package internal

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	shareddomain "zensor-server/internal/shared_kernel/domain"
)

type Activity struct {
	ID                     string      `json:"id" gorm:"primaryKey"`
	Version                int         `json:"version"`
	TenantID               string      `json:"tenant_id" gorm:"index;not null"`
	TypeName               string      `json:"type_name" gorm:"not null"`
	CustomTypeName         *string     `json:"custom_type_name,omitempty"`
	Name                   string      `json:"name" gorm:"not null"`
	Description            string      `json:"description"`
	Schedule               string      `json:"schedule" gorm:"not null"`
	NotificationDaysBefore Days        `json:"notification_days_before"`
	Fields                 Fields      `json:"fields"`
	IsActive               bool        `json:"is_active" gorm:"default:true"`
	CreatedAt              utils.Time  `json:"created_at"`
	UpdatedAt              utils.Time  `json:"updated_at"`
	DeletedAt              *utils.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Activity) TableName() string {
	return "maintenance_activities"
}

type Days []int

func (d Days) Value() (driver.Value, error) {
	if len(d) == 0 {
		return "[]", nil
	}
	return json.Marshal(d)
}

func (d *Days) Scan(src any) error {
	var data []byte

	switch val := src.(type) {
	case string:
		data = []byte(val)
	case []byte:
		data = val
	case nil:
		*d = Days{}
		return nil
	default:
		return errors.New("invalid type for days")
	}

	return json.Unmarshal(data, d)
}

type Fields []maintenanceDomain.FieldDefinition

func (f Fields) Value() (driver.Value, error) {
	if len(f) == 0 {
		return "[]", nil
	}
	return json.Marshal(f)
}

func (f *Fields) Scan(src any) error {
	var data []byte

	switch val := src.(type) {
	case string:
		data = []byte(val)
	case []byte:
		data = val
	case nil:
		*f = Fields{}
		return nil
	default:
		return errors.New("invalid type for fields")
	}

	return json.Unmarshal(data, f)
}

func (m Activity) ToDomain() maintenanceDomain.Activity {
	result := maintenanceDomain.Activity{
		ID:                     shareddomain.ID(m.ID),
		Version:                shareddomain.Version(m.Version),
		TenantID:               shareddomain.ID(m.TenantID),
		Name:                   shareddomain.Name(m.Name),
		Description:            shareddomain.Description(m.Description),
		Schedule:               maintenanceDomain.Schedule(m.Schedule),
		NotificationDaysBefore: maintenanceDomain.Days(m.NotificationDaysBefore),
		Fields:                 []maintenanceDomain.FieldDefinition(m.Fields),
		IsActive:               m.IsActive,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}

	result.Type = maintenanceDomain.ActivityType{
		ID:           shareddomain.ID(m.TypeName),
		Name:         shareddomain.Name(m.TypeName),
		DisplayName:  shareddomain.DisplayName(""),
		Description:  shareddomain.Description(""),
		IsPredefined: true,
	}

	if m.CustomTypeName != nil {
		customTypeName := maintenanceDomain.CustomTypeName(*m.CustomTypeName)
		result.CustomTypeName = &customTypeName
	}

	if m.DeletedAt != nil {
		result.DeletedAt = m.DeletedAt
	}

	return result
}

func FromActivity(value maintenanceDomain.Activity) Activity {
	result := Activity{
		ID:                     value.ID.String(),
		Version:                int(value.Version),
		TenantID:               value.TenantID.String(),
		TypeName:               string(value.Type.Name),
		Name:                   string(value.Name),
		Description:            string(value.Description),
		Schedule:               string(value.Schedule),
		NotificationDaysBefore: Days(value.NotificationDaysBefore),
		Fields:                 Fields(value.Fields),
		IsActive:               value.IsActive,
		CreatedAt:              value.CreatedAt,
		UpdatedAt:              value.UpdatedAt,
	}

	if value.CustomTypeName != nil {
		customTypeName := string(*value.CustomTypeName)
		result.CustomTypeName = &customTypeName
	}

	if value.DeletedAt != nil {
		result.DeletedAt = value.DeletedAt
	}

	return result
}
