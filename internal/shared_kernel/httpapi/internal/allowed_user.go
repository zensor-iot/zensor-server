package internal

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

// AllowedUserResponse represents an allowlist entry in admin API responses
type AllowedUserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name,omitempty"`
	IsAdmin     bool       `json:"is_admin"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// AllowedUserCreateRequest represents the request for adding an allowlist entry
type AllowedUserCreateRequest struct {
	Email   string `json:"email" validate:"required"`
	IsAdmin bool   `json:"is_admin"`
}

// AllowedUserUpdateRequest represents the request for updating an allowlist entry
type AllowedUserUpdateRequest struct {
	IsAdmin bool `json:"is_admin"`
}

// CurrentUserResponse represents the authenticated user returned by /v1/me
type CurrentUserResponse struct {
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

// ToAllowedUserResponse converts a domain.AllowedUser to AllowedUserResponse
func ToAllowedUserResponse(user domain.AllowedUser) AllowedUserResponse {
	return AllowedUserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		DisplayName: user.DisplayName,
		IsAdmin:     user.IsAdmin,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// ToCurrentUserResponse converts a domain.Session to CurrentUserResponse
func ToCurrentUserResponse(session domain.Session) CurrentUserResponse {
	return CurrentUserResponse{
		UserID:  session.UserID.String(),
		Name:    session.Name,
		Email:   session.Email,
		IsAdmin: session.IsAdmin,
	}
}
