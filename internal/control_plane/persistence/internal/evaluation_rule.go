package internal

import (
	"time"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/utils"
)

type EvaluationRule struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	Device      Device         `json:"device" gorm:"foreignKey:device_id"`
	Version     int            `json:"version"`
	Description string         `json:"description"`
	Kind        string         `json:"kind"`
	Enabled     bool           `json:"enabled"`
	Parameters  map[string]any `json:"parameters"`
	CreatedAt   utils.Time     `json:"created_at"`
	UpdaatedAt  utils.Time     `json:"updated_at"`
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
		UpdaatedAt:  utils.Time{Time: time.Now()},
	}
}

func MapEvaluationRuleParametersToMap(parameters []domain.EvaluationRuleParameter) map[string]any {
	result := make(map[string]any)
	for _, parameter := range parameters {
		result[parameter.Key] = parameter.Value
	}
	return result
}
