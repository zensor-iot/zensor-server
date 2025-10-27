package internal

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"
	"zensor-server/internal/shared_kernel/domain"
)

type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Tenants   TenantIDs `json:"tenants" gorm:"type:text[]"`
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
