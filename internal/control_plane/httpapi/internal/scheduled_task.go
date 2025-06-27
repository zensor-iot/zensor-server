package internal

type ScheduledTaskCreateRequest struct {
	Commands []CommandSendPayloadRequest `json:"commands"`
	Schedule string                      `json:"schedule"`
	IsActive bool                        `json:"is_active"`
}

type ScheduledTaskUpdateRequest struct {
	Commands *[]CommandSendPayloadRequest `json:"commands,omitempty"`
	Schedule *string                      `json:"schedule,omitempty"`
	IsActive *bool                        `json:"is_active,omitempty"`
}

type ScheduledTaskResponse struct {
	ID       string                      `json:"id"`
	DeviceID string                      `json:"device_id"`
	Commands []CommandSendPayloadRequest `json:"commands"`
	Schedule string                      `json:"schedule"`
	IsActive bool                        `json:"is_active"`
}

type ScheduledTaskListResponse struct {
	ScheduledTasks []ScheduledTaskResponse `json:"scheduled_tasks"`
}
