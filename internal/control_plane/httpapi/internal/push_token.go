package internal

type PushTokenRegistrationRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

type PushTokenUnregistrationRequest struct {
	Token string `json:"token"`
}

