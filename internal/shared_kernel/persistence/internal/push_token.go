package internal

import (
	"time"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"
)

type PushToken struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"index;not null"`
	Token     string    `json:"token" gorm:"not null;uniqueIndex"`
	Platform  string    `json:"platform" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PushToken) TableName() string {
	return "push_tokens"
}

func (pt PushToken) ToDomain() domain.PushToken {
	return domain.PushToken{
		ID:        domain.ID(pt.ID),
		UserID:    domain.ID(pt.UserID),
		Token:     pt.Token,
		Platform:  pt.Platform,
		CreatedAt: utils.Time{Time: pt.CreatedAt},
		UpdatedAt: utils.Time{Time: pt.UpdatedAt},
	}
}

func FromPushToken(value domain.PushToken) PushToken {
	return PushToken{
		ID:        value.ID.String(),
		UserID:    value.UserID.String(),
		Token:     value.Token,
		Platform:  value.Platform,
		CreatedAt: value.CreatedAt.Time,
		UpdatedAt: value.UpdatedAt.Time,
	}
}
