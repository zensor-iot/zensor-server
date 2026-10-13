package internal

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

// AuthModeResponse tells the frontend which login flow to render.
type AuthModeResponse struct {
	Mode string `json:"mode"`
}

// StaticLoginRequest carries the credentials submitted to the static-mode login endpoint.
type StaticLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// MobileExchangeRequest carries the single-use code the app got in the login redirect.
type MobileExchangeRequest struct {
	Code string `json:"code"`
}

// MobileSessionResponse is the bearer session the app sends as
// `Authorization: Bearer <token>`.
type MobileSessionResponse struct {
	Token     string              `json:"token"`
	ExpiresAt time.Time           `json:"expires_at"`
	User      CurrentUserResponse `json:"user"`
}

func ToMobileSessionResponse(session domain.Session) MobileSessionResponse {
	return MobileSessionResponse{
		Token:     session.ID,
		ExpiresAt: session.ExpiresAt,
		User:      ToCurrentUserResponse(session),
	}
}
