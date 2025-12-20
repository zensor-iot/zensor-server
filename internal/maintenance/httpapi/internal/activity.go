package internal

import (
	"time"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
)

type ActivityListResponse struct {
	Data []ActivityResponse `json:"data"`
}

type ActivityResponse struct {
	ID                     string                    `json:"id"`
	Version                int                       `json:"version"`
	TenantID               string                    `json:"tenant_id"`
	TypeName               string                    `json:"type_name"`
	CustomTypeName         *string                   `json:"custom_type_name,omitempty"`
	Name                   string                    `json:"name"`
	Description            string                    `json:"description"`
	Schedule               string                    `json:"schedule"`
	NotificationDaysBefore []int                     `json:"notification_days_before"`
	Fields                 []FieldDefinitionResponse `json:"fields"`
	IsActive               bool                      `json:"is_active"`
	CreatedAt              time.Time                 `json:"created_at"`
	UpdatedAt              time.Time                 `json:"updated_at"`
}

type FieldDefinitionResponse struct {
	Name         string  `json:"name"`
	DisplayName  string  `json:"display_name"`
	Type         string  `json:"type"`
	IsRequired   bool    `json:"is_required"`
	DefaultValue *string `json:"default_value,omitempty"`
}

type ActivityCreateRequest struct {
	TenantID               string                   `json:"tenant_id"`
	TypeName               string                   `json:"type_name"`
	CustomTypeName         *string                  `json:"custom_type_name,omitempty"`
	Name                   string                   `json:"name"`
	Description            string                   `json:"description"`
	Schedule               string                   `json:"schedule"`
	NotificationDaysBefore []int                    `json:"notification_days_before"`
	Fields                 []FieldDefinitionRequest `json:"fields"`
}

type FieldDefinitionRequest struct {
	Name         string  `json:"name"`
	DisplayName  string  `json:"display_name"`
	Type         string  `json:"type"`
	IsRequired   bool    `json:"is_required"`
	DefaultValue *string `json:"default_value,omitempty"`
}

type ActivityUpdateRequest struct {
	Name                   *string                   `json:"name,omitempty"`
	Description            *string                   `json:"description,omitempty"`
	Schedule               *string                   `json:"schedule,omitempty"`
	NotificationDaysBefore *[]int                    `json:"notification_days_before,omitempty"`
	Fields                 *[]FieldDefinitionRequest `json:"fields,omitempty"`
}

func ToActivityResponse(activity maintenanceDomain.Activity) ActivityResponse {
	response := ActivityResponse{
		ID:                     activity.ID.String(),
		Version:                int(activity.Version),
		TenantID:               activity.TenantID.String(),
		TypeName:               string(activity.Type.Name),
		Name:                   string(activity.Name),
		Description:            string(activity.Description),
		Schedule:               string(activity.Schedule),
		NotificationDaysBefore: []int(activity.NotificationDaysBefore),
		Fields:                 []FieldDefinitionResponse{},
		IsActive:               activity.IsActive,
		CreatedAt:              activity.CreatedAt.Time,
		UpdatedAt:              activity.UpdatedAt.Time,
	}

	if activity.CustomTypeName != nil {
		customTypeName := string(*activity.CustomTypeName)
		response.CustomTypeName = &customTypeName
	}

	// Convert fields
	for _, field := range activity.Fields {
		fieldResp := FieldDefinitionResponse{
			Name:        string(field.Name),
			DisplayName: string(field.DisplayName),
			Type:        string(field.Type),
			IsRequired:  field.IsRequired,
		}
		if field.DefaultValue != nil {
			// Convert any to string
			if str, ok := (*field.DefaultValue).(string); ok {
				fieldResp.DefaultValue = &str
			}
		}
		response.Fields = append(response.Fields, fieldResp)
	}

	return response
}

func ToActivityListResponse(activities []maintenanceDomain.Activity) ActivityListResponse {
	responses := make([]ActivityResponse, len(activities))
	for i, activity := range activities {
		responses[i] = ToActivityResponse(activity)
	}

	return ActivityListResponse{
		Data: responses,
	}
}
