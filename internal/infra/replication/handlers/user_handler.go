package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
)

type UserData struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Tenants   string    `json:"tenants" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserData) TableName() string {
	return "users"
}

type UserHandler struct {
	orm sql.ORM
}

func NewUserHandler(orm sql.ORM) *UserHandler {
	return &UserHandler{
		orm: orm,
	}
}

func (h *UserHandler) TopicName() pubsub.Topic {
	return "users"
}

func (h *UserHandler) Create(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	userData := h.extractUserFields(message)

	err := h.orm.WithContext(ctx).Create(&userData).Error()
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	return nil
}

func (h *UserHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var userData UserData

	err := h.orm.WithContext(ctx).First(&userData, "id = ?", id).Error()
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	user := h.toAvroUser(userData)
	return user, nil
}

func (h *UserHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	userData := h.extractUserFields(message)

	err := h.orm.WithContext(ctx).Save(&userData).Error()
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}

	return nil
}

func (h *UserHandler) extractUserFields(msg pubsub.Message) UserData {
	avroUser, ok := msg.(*avro.AvroUser)
	if !ok {
		panic("invalid message type for UserHandler")
	}

	return UserData{
		ID:        avroUser.ID,
		Tenants:   marshallTenants(avroUser.Tenants),
		CreatedAt: avroUser.CreatedAt,
		UpdatedAt: avroUser.UpdatedAt,
	}
}

func (h *UserHandler) toAvroUser(data UserData) *avro.AvroUser {
	return &avro.AvroUser{
		ID:        data.ID,
		Tenants:   unmarshallTenants(data.Tenants),
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.UpdatedAt,
	}
}

func marshallTenants(tenants []string) string {
	if len(tenants) == 0 {
		return "[]"
	}
	bytes, _ := json.Marshal(tenants)
	return string(bytes)
}

func unmarshallTenants(tenantsJSON string) []string {
	var tenants []string
	json.Unmarshal([]byte(tenantsJSON), &tenants)
	return tenants
}
