package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"zensor-server/internal/infra/utils"
)

// APIKeyPlaintextPrefix is the fixed prefix identifying Zensor API keys.
const APIKeyPlaintextPrefix = "zsk_"

const apiKeyRandomBytes = 32

const apiKeyPrefixLength = 12

// ErrAPIKeyNameRequired signals a builder attempt without a usable name.
var ErrAPIKeyNameRequired = errors.New("api key name is required")

// APIKey is a machine credential granting non-human clients API access.
// Only the SHA-256 hash of the plaintext key is retained.
type APIKey struct {
	ID        ID
	Name      string
	KeyHash   string
	KeyPrefix string
	CreatedBy ID
	CreatedAt time.Time
}

// HashAPIKey returns the SHA-256 hex digest used to store and look up keys.
func HashAPIKey(plaintext string) string {
	digest := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(digest[:])
}

func NewAPIKeyBuilder() *apiKeyBuilder {
	return &apiKeyBuilder{}
}

type apiKeyBuilder struct {
	actions []apiKeyHandler
}

type apiKeyHandler func(k *APIKey) error

func (b *apiKeyBuilder) WithName(name string) *apiKeyBuilder {
	b.actions = append(b.actions, func(k *APIKey) error {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return ErrAPIKeyNameRequired
		}
		k.Name = trimmed
		return nil
	})
	return b
}

func (b *apiKeyBuilder) WithCreatedBy(createdBy ID) *apiKeyBuilder {
	b.actions = append(b.actions, func(k *APIKey) error {
		k.CreatedBy = createdBy
		return nil
	})
	return b
}

// Build assembles the API key and returns it together with the plaintext
// key, which is available exactly once at creation time.
func (b *apiKeyBuilder) Build() (APIKey, string, error) {
	result := APIKey{
		ID:        ID(utils.GenerateUUID()),
		CreatedAt: time.Now(),
	}

	for _, action := range b.actions {
		if err := action(&result); err != nil {
			return APIKey{}, "", err
		}
	}

	if result.Name == "" {
		return APIKey{}, "", ErrAPIKeyNameRequired
	}

	plaintext, err := generateAPIKeyPlaintext()
	if err != nil {
		return APIKey{}, "", err
	}

	result.KeyHash = HashAPIKey(plaintext)
	result.KeyPrefix = plaintext[:apiKeyPrefixLength]

	return result, plaintext, nil
}

func generateAPIKeyPlaintext() (string, error) {
	buffer := make([]byte, apiKeyRandomBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("reading random bytes: %w", err)
	}

	return APIKeyPlaintextPrefix + hex.EncodeToString(buffer), nil
}
