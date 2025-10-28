package internal

import (
	"zensor-server/internal/shared_kernel/domain"
)

type UserResponse struct {
	ID      string   `json:"id"`
	Tenants []string `json:"tenants"`
}

type UserUpdateRequest struct {
	Tenants []string `json:"tenants"`
}

func ToUserResponse(user domain.User) UserResponse {
	tenantIDs := make([]string, len(user.Tenants))
	for i, tenantID := range user.Tenants {
		tenantIDs[i] = tenantID.String()
	}

	return UserResponse{
		ID:      user.ID.String(),
		Tenants: tenantIDs,
	}
}
