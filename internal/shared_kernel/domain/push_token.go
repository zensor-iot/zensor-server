package domain

import (
	"errors"
	"time"
	"zensor-server/internal/infra/utils"
)

type PushToken struct {
	ID        ID
	UserID    ID
	Token     string
	Platform  string
	CreatedAt utils.Time
	UpdatedAt utils.Time
}

func NewPushTokenBuilder() *pushTokenBuilder {
	return &pushTokenBuilder{}
}

type pushTokenBuilder struct {
	actions []pushTokenHandler
}

type pushTokenHandler func(v *PushToken) error

func (b *pushTokenBuilder) WithUserID(value ID) *pushTokenBuilder {
	b.actions = append(b.actions, func(d *PushToken) error {
		d.UserID = value
		return nil
	})
	return b
}

func (b *pushTokenBuilder) WithToken(value string) *pushTokenBuilder {
	b.actions = append(b.actions, func(d *PushToken) error {
		d.Token = value
		return nil
	})
	return b
}

func (b *pushTokenBuilder) WithPlatform(value string) *pushTokenBuilder {
	b.actions = append(b.actions, func(d *PushToken) error {
		d.Platform = value
		return nil
	})
	return b
}

func (b *pushTokenBuilder) Build() (PushToken, error) {
	now := utils.Time{Time: time.Now()}
	result := PushToken{
		ID:        ID(utils.GenerateUUID()),
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return PushToken{}, err
		}
	}

	if result.UserID == "" {
		return PushToken{}, errors.New("user ID is required")
	}

	if result.Token == "" {
		return PushToken{}, errors.New("token is required")
	}

	if result.Platform == "" {
		return PushToken{}, errors.New("platform is required")
	}

	return result, nil
}
