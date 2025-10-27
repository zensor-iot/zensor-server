package internal

import (
	"database/sql/driver"
	"encoding/json"
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Tenants   TenantIDs `json:"tenants" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

func (s User) ToDomain() domain.User {
	return domain.User{
		ID:      domain.ID(s.ID),
		Tenants: s.Tenants.ToDomainIDs(),
	}
}

func FromUser(value domain.User) User {
	return User{
		ID:        value.ID.String(),
		Tenants:   TenantIDsFromDomainIDs(value.Tenants),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

type TenantIDs []string

func (t *TenantIDs) ToDomainIDs() []domain.ID {
	result := make([]domain.ID, len(*t))
	for i, id := range *t {
		result[i] = domain.ID(id)
	}
	return result
}

func TenantIDsFromDomainIDs(ids []domain.ID) TenantIDs {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}

func (t *TenantIDs) Scan(value interface{}) error {
	if value == nil {
		*t = TenantIDs{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, t)
}

func (t TenantIDs) Value() (driver.Value, error) {
	if len(t) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(t)
}
