package domain

import (
	"strings"
	"time"
	"zensor-server/internal/infra/utils"
)

// AllowedUser is an allowlist entry granting a Google account access to the portal.
type AllowedUser struct {
	ID          ID
	Email       string
	DisplayName string
	IsAdmin     bool
	LastLoginAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RecordLogin captures the identity details observed during a successful login.
func (u *AllowedUser) RecordLogin(displayName string) {
	now := time.Now()
	u.DisplayName = displayName
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// NormalizeEmail lowercases and trims an email address for allowlist comparison.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NewAllowedUserBuilder() *allowedUserBuilder {
	return &allowedUserBuilder{}
}

type allowedUserBuilder struct {
	actions []allowedUserHandler
}

type allowedUserHandler func(u *AllowedUser) error

func (b *allowedUserBuilder) WithEmail(email string) *allowedUserBuilder {
	b.actions = append(b.actions, func(u *AllowedUser) error {
		normalized := NormalizeEmail(email)
		if err := utils.ValidateEmail(normalized); err != nil {
			return err
		}
		u.Email = normalized
		return nil
	})
	return b
}

func (b *allowedUserBuilder) WithDisplayName(displayName string) *allowedUserBuilder {
	b.actions = append(b.actions, func(u *AllowedUser) error {
		u.DisplayName = displayName
		return nil
	})
	return b
}

func (b *allowedUserBuilder) WithIsAdmin(isAdmin bool) *allowedUserBuilder {
	b.actions = append(b.actions, func(u *AllowedUser) error {
		u.IsAdmin = isAdmin
		return nil
	})
	return b
}

func (b *allowedUserBuilder) Build() (AllowedUser, error) {
	now := time.Now()
	result := AllowedUser{
		ID:        ID(utils.GenerateUUID()),
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, action := range b.actions {
		if err := action(&result); err != nil {
			return AllowedUser{}, err
		}
	}

	if err := utils.ValidateEmail(result.Email); err != nil {
		return AllowedUser{}, err
	}

	return result, nil
}
