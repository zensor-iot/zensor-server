package internal

type PushTokenRegistrationRequest struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

type PushTokenUnregistrationRequest struct {
	Token string `json:"token"`
}

type UserPushBroadcastRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	DeepLink string `json:"deep_link"`
}
