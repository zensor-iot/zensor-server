package internal

import (
	"time"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
)

type ExecutionListResponse struct {
	Data []ExecutionResponse `json:"data"`
}

type ExecutionResponse struct {
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

type ExecutionCreateRequest struct {
	ActivityID    string                 `json:"activity_id"`
	ScheduledDate time.Time              `json:"scheduled_date"`
	FieldValues   map[string]interface{} `json:"field_values"`
}

type ExecutionCompleteRequest struct {
	CompletedBy string `json:"completed_by"`
}

func ToExecutionResponse(execution maintenanceDomain.Execution) ExecutionResponse {
	response := ExecutionResponse{
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

func ToExecutionListResponse(executions []maintenanceDomain.Execution) ExecutionListResponse {
	responses := make([]ExecutionResponse, len(executions))
	for i, execution := range executions {
		responses[i] = ToExecutionResponse(execution)
	}

	return ExecutionListResponse{
		Data: responses,
	}
}
