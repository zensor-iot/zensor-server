package internal

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

type APIKey struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex;not null"`
	KeyHash   string    `json:"key_hash" gorm:"uniqueIndex;not null"`
	KeyPrefix string    `json:"key_prefix"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

func (k APIKey) ToDomain() domain.APIKey {
	return domain.APIKey{
		ID:        domain.ID(k.ID),
		Name:      k.Name,
		KeyHash:   k.KeyHash,
		KeyPrefix: k.KeyPrefix,
		CreatedBy: domain.ID(k.CreatedBy),
		CreatedAt: k.CreatedAt,
	}
}

func FromAPIKey(value domain.APIKey) APIKey {
	return APIKey{
		ID:        value.ID.String(),
		Name:      value.Name,
		KeyHash:   value.KeyHash,
		KeyPrefix: value.KeyPrefix,
		CreatedBy: value.CreatedBy.String(),
		CreatedAt: value.CreatedAt,
	}
}
