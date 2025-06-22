package internal

type ScheduledTaskCreateRequest struct {
	DeviceID string            `json:"device_id"`
	Task     TaskCreateRequest `json:"task"`
	Schedule string            `json:"schedule"`
	IsActive bool              `json:"is_active"`
}

type ScheduledTaskUpdateRequest struct {
	Task     *TaskCreateRequest `json:"task,omitempty"`
	Schedule *string            `json:"schedule,omitempty"`
	IsActive *bool              `json:"is_active,omitempty"`
}

type ScheduledTaskResponse struct {
	ID       string `json:"id"`
	DeviceID string `json:"device_id"`
	TaskID   string `json:"task_id"`
	Schedule string `json:"schedule"`
	IsActive bool   `json:"is_active"`
}

type ScheduledTaskListResponse struct {
	ScheduledTasks []ScheduledTaskResponse `json:"scheduled_tasks"`
}
