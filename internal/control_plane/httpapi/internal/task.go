package internal

type TaskCreateRequest struct {
	Commands []CommandSendPayloadRequest `json:"commands"`
}

type TaskResponse struct {
	ID string `json:"id"`
}

type TaskListResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}
