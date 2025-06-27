package internal

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"
)

type EvaluationRule struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	DeviceID    string     `json:"device_id" gorm:"foreignKey:device_id"`
	Version     int        `json:"version"`
	Description string     `json:"description"`
	Kind        string     `json:"kind"`
	Enabled     bool       `json:"enabled"`
	Parameters  Parameters `json:"parameters"`
	CreatedAt   utils.Time `json:"created_at"`
	UpdatedAt   utils.Time `json:"updated_at"`
}

func (EvaluationRule) TableName() string {
	return "evaluation_rules_final"
}

type Parameters map[string]any

func (p Parameters) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Parameters) Scan(src any) error {
	data, ok := src.(string)
	if !ok {
		return errors.New("invalid type")
	}
	return json.Unmarshal([]byte(data), p)
}

func FromEvaluationRule(value domain.EvaluationRule) EvaluationRule {
	return EvaluationRule{
		ID:          value.ID.String(),
		Version:     1,
		Description: value.Description,
		Kind:        value.Kind,
		Enabled:     value.Enabled,
		Parameters:  MapEvaluationRuleParametersToMap(value.Parameters),
		CreatedAt:   utils.Time{Time: time.Now()},
		UpdatedAt:   utils.Time{Time: time.Now()},
	}
}

func MapEvaluationRuleParametersToMap(parameters []domain.EvaluationRuleParameter) map[string]any {
	result := make(map[string]any)
	for _, parameter := range parameters {
		result[parameter.Key] = parameter.Value
	}
	return result
}

func (e EvaluationRule) ToDomain() domain.EvaluationRule {
	return domain.EvaluationRule{
		ID:          domain.ID(e.ID),
		Version:     domain.Version(e.Version),
		Description: e.Description,
		Kind:        e.Kind,
		Enabled:     e.Enabled,
		Parameters:  MapToEvaluationRuleParameters(e.Parameters),
	}
}

func MapToEvaluationRuleParameters(parameters map[string]any) []domain.EvaluationRuleParameter {
	var result []domain.EvaluationRuleParameter
	for key, value := range parameters {
		result = append(result, domain.EvaluationRuleParameter{
			Key:   key,
			Value: value,
		})
	}
	return result
}
