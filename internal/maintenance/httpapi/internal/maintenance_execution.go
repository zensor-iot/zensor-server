package internal

import (
	"time"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
)

type MaintenanceExecutionListResponse struct {
	Data []MaintenanceExecutionResponse `json:"data"`
}

type MaintenanceExecutionResponse struct {
	ID            string                 `json:"id"`
	Version       int                    `json:"version"`
	ActivityID    string                 `json:"activity_id"`
	ScheduledDate time.Time              `json:"scheduled_date"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	CompletedBy   *string                `json:"completed_by,omitempty"`
	IsOverdue     bool                   `json:"is_overdue"`
	OverdueDays   int                    `json:"overdue_days"`
	FieldValues   map[string]interface{} `json:"field_values"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type MaintenanceExecutionCreateRequest struct {
	ActivityID    string                 `json:"activity_id"`
	ScheduledDate time.Time              `json:"scheduled_date"`
	FieldValues   map[string]interface{} `json:"field_values"`
}

type MaintenanceExecutionCompleteRequest struct {
	CompletedBy string `json:"completed_by"`
}

func ToMaintenanceExecutionResponse(execution maintenanceDomain.MaintenanceExecution) MaintenanceExecutionResponse {
	response := MaintenanceExecutionResponse{
		ID:            execution.ID.String(),
		Version:       int(execution.Version),
		ActivityID:    execution.ActivityID.String(),
		ScheduledDate: execution.ScheduledDate.Time,
		IsOverdue:     execution.IsOverdue(),
		OverdueDays:   int(execution.OverdueDays),
		FieldValues:   execution.FieldValues,
		CreatedAt:     execution.CreatedAt.Time,
		UpdatedAt:     execution.UpdatedAt.Time,
	}

	if execution.CompletedAt != nil {
		response.CompletedAt = &execution.CompletedAt.Time
	}

	if execution.CompletedBy != nil {
		completedBy := string(*execution.CompletedBy)
		response.CompletedBy = &completedBy
	}

	return response
}

func ToMaintenanceExecutionListResponse(executions []maintenanceDomain.MaintenanceExecution) MaintenanceExecutionListResponse {
	responses := make([]MaintenanceExecutionResponse, len(executions))
	for i, execution := range executions {
		responses[i] = ToMaintenanceExecutionResponse(execution)
	}

	return MaintenanceExecutionListResponse{
		Data: responses,
	}
}
