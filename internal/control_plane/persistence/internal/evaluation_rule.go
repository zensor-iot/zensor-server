package internal

import (
	"time"
)

type EvaluationRule struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	Version     int            `json:"version"`
	Description string         `json:"description"`
	Kind        string         `json:"kind"`
	Enabled     bool           `json:"enabled"`
	Parameters  map[string]any `json:"parameters"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdaatedAt  time.Time      `json:"updated_at"`
}

// func FromEvaluationRule(value domain.EvaluationRule) EvaluationRule {
// 	return EvaluationRule{
// 		ID:          value.ID.String(),
// 		Version:     1,
// 		Description: value.Description,
// 		Kind:        value.Kind,
// 		Enabled:     value.Enabled,
// 		CreatedAt:   time.Now(),
// 		UpdaatedAt:  time.Now(),
// 	}
// }

// func MapEvaluationRuleParametersToMap(parameters []domain.Eva) map[string]any {
// 	return utils.Map(parameters, func(key string, value any) (string, any) {
// 		return key, value
// 	})
// }
