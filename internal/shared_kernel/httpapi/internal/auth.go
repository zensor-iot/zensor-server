package internal

// AuthModeResponse tells the frontend which login flow to render.
type AuthModeResponse struct {
	Mode string `json:"mode"`
}

// StaticLoginRequest carries the credentials submitted to the static-mode login endpoint.
type StaticLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
