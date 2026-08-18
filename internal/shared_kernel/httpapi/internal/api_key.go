package internal

import (
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

// APIKeyCreateRequest represents the request for creating an API key.
type APIKeyCreateRequest struct {
	Name string `json:"name" validate:"required"`
}

// APIKeyCreatedResponse is the create response carrying the plaintext key,
// which is exposed exactly once at creation time.
type APIKeyCreatedResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	KeyPrefix string    `json:"key_prefix"`
	CreatedAt time.Time `json:"created_at"`
}

// APIKeyResponse represents an API key in listings, never exposing key material.
type APIKeyResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	KeyPrefix string    `json:"key_prefix"`
	CreatedAt time.Time `json:"created_at"`
}

// ToAPIKeyCreatedResponse converts a freshly created key and its plaintext to the create response.
func ToAPIKeyCreatedResponse(key domain.APIKey, plaintext string) APIKeyCreatedResponse {
	return APIKeyCreatedResponse{
		ID:        key.ID.String(),
		Name:      key.Name,
		Key:       plaintext,
		KeyPrefix: key.KeyPrefix,
		CreatedAt: key.CreatedAt,
	}
}

// ToAPIKeyResponse converts a domain.APIKey to APIKeyResponse.
func ToAPIKeyResponse(key domain.APIKey) APIKeyResponse {
	return APIKeyResponse{
		ID:        key.ID.String(),
		Name:      key.Name,
		KeyPrefix: key.KeyPrefix,
		CreatedAt: key.CreatedAt,
	}
}
