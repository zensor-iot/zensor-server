package handlers

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
)

type UserData struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Tenants   TenantIDs `json:"tenants" gorm:"type:text[]"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserData) TableName() string {
	return "users"
}

type TenantIDs []string

func (t *TenantIDs) Scan(value interface{}) error {
	if value == nil {
		*t = TenantIDs{}
		return nil
	}

	switch v := value.(type) {
	case string:
		*t = parsePostgresArray(v)
		return nil
	case []byte:
		result := parsePostgresArray(string(v))
		*t = result
		return nil
	case []interface{}:
		result := make(TenantIDs, len(v))
		for i, item := range v {
			switch val := item.(type) {
			case string:
				result[i] = val
			case []interface{}:
				if len(val) > 0 {
					if str, ok := val[0].(string); ok {
						result[i] = str
					}
				}
			}
		}
		*t = result
		return nil
	default:
		bytes, ok := value.([]byte)
		if !ok {
			*t = TenantIDs{}
			return nil
		}
		return json.Unmarshal(bytes, t)
	}
}

func (t TenantIDs) Value() (driver.Value, error) {
	if len(t) == 0 {
		return "{}", nil
	}
	result := "{"
	for i, tenant := range t {
		if i > 0 {
			result += ","
		}
		result += tenant
	}
	result += "}"
	return result, nil
}

func parsePostgresArray(s string) TenantIDs {
	s = strings.Trim(s, "{}")
	if s == "" {
		return TenantIDs{}
	}

	parts := strings.Split(s, ",")
	result := make(TenantIDs, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}
	return result
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
	avroUser := h.extractAvroUser(message)
	userData := UserData{
		ID:        avroUser.ID,
		Tenants:   avroUser.Tenants,
		CreatedAt: avroUser.CreatedAt,
		UpdatedAt: avroUser.UpdatedAt,
	}

	err := h.orm.WithContext(ctx).Create(&userData).Error()
	if err != nil {
		slog.Error("failed to create user in database", slog.String("error", err.Error()))
		return fmt.Errorf("creating user: %w", err)
	}

	slog.Info("successfully created user", slog.String("key", string(key)), slog.Any("tenants", userData.Tenants))
	return nil
}

func (h *UserHandler) GetByID(ctx context.Context, id string) (pubsub.Message, error) {
	var userData UserData

	err := h.orm.WithContext(ctx).First(&userData, "id = ?", id).Error()
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}

	avroUser := &avro.AvroUser{
		ID:        userData.ID,
		Tenants:   userData.Tenants,
		CreatedAt: userData.CreatedAt,
		UpdatedAt: userData.UpdatedAt,
	}
	return avroUser, nil
}

func (h *UserHandler) Update(ctx context.Context, key pubsub.Key, message pubsub.Message) error {
	avroUser := h.extractAvroUser(message)
	userData := UserData{
		ID:        avroUser.ID,
		Tenants:   avroUser.Tenants,
		CreatedAt: avroUser.CreatedAt,
		UpdatedAt: avroUser.UpdatedAt,
	}

	err := h.orm.WithContext(ctx).Save(&userData).Error()
	if err != nil {
		slog.Error("failed to update user in database", slog.String("error", err.Error()))
		return fmt.Errorf("updating user: %w", err)
	}

	slog.Info("successfully updated user", slog.String("key", string(key)), slog.Any("tenants", userData.Tenants))
	return nil
}

func (h *UserHandler) extractAvroUser(msg pubsub.Message) *avro.AvroUser {
	avroUser, ok := msg.(*avro.AvroUser)
	if !ok {
		slog.Error("invalid message type for UserHandler", slog.String("type", fmt.Sprintf("%T", msg)))
		panic("invalid message type for UserHandler")
	}
	slog.Debug("extracting user fields", slog.String("user_id", avroUser.ID))
	return avroUser
}
