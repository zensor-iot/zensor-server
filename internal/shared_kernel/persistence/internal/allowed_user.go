package internal

import (
	"time"

	"zensor-server/internal/shared_kernel/domain"
)

type AllowedUser struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Email       string     `json:"email" gorm:"uniqueIndex;not null"`
	DisplayName string     `json:"display_name"`
	IsAdmin     bool       `json:"is_admin"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (AllowedUser) TableName() string {
	return "allowed_users"
}

func (u AllowedUser) ToDomain() domain.AllowedUser {
	return domain.AllowedUser{
		ID:          domain.ID(u.ID),
		Email:       u.Email,
		DisplayName: u.DisplayName,
		IsAdmin:     u.IsAdmin,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func FromAllowedUser(value domain.AllowedUser) AllowedUser {
	return AllowedUser{
		ID:          value.ID.String(),
		Email:       value.Email,
		DisplayName: value.DisplayName,
		IsAdmin:     value.IsAdmin,
		LastLoginAt: value.LastLoginAt,
		CreatedAt:   value.CreatedAt,
		UpdatedAt:   value.UpdatedAt,
	}
}
