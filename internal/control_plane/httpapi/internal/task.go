package internal

type TaskCreateRequest struct {
	Commands []CommandSendPayloadRequest `json:"commands"`
}

type TaskCommandResponse struct {
	ID            string  `json:"id"`
	Index         uint8   `json:"index"`
	Value         uint8   `json:"value"`
	Port          uint8   `json:"port"`
	Priority      string  `json:"priority"`
	DispatchAfter string  `json:"dispatch_after"`
	Ready         bool    `json:"ready"`
	Sent          bool    `json:"sent"`
	SentAt        *string `json:"sent_at,omitempty"`
}

type TaskResponse struct {
	ID       string                `json:"id"`
	Commands []TaskCommandResponse `json:"commands"`
}

type TaskListResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}
