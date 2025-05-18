package domain

import (
	"errors"
	"zensor-server/internal/infra/utils"
)

type EvaluationRule struct {
	ID          ID
	Description string
	Kind        string
	Parameters  []EvaluetionRuleParameter
	Enabled     bool
}

func (er *EvaluationRule) AddParameters(params ...EvaluetionRuleParameter) {
	er.Parameters = append(er.Parameters, params...)
}

type EvaluetionRuleParameter struct {
	Key   string
	Value any
}

func NewEvaluationRuleBuilder() *evaluationRuleBuilder {
	return &evaluationRuleBuilder{}
}

type evaluationRuleBuilder struct {
	actions []evaluationRuleHandler
}

type evaluationRuleHandler func(v *EvaluationRule) error

func (b *evaluationRuleBuilder) WithDescription(value string) *evaluationRuleBuilder {
	b.actions = append(b.actions, func(d *EvaluationRule) error {
		d.Description = value
		return nil
	})
	return b
}

func (b *evaluationRuleBuilder) WithKind(value string) *evaluationRuleBuilder {
	b.actions = append(b.actions, func(d *EvaluationRule) error {
		d.Kind = value
		return nil
	})
	return b
}

func (b *evaluationRuleBuilder) WithParameters(value ...EvaluetionRuleParameter) *evaluationRuleBuilder {
	b.actions = append(b.actions, func(d *EvaluationRule) error {
		d.Parameters = value
		return nil
	})
	return b
}

var (
	ErrInvalidThresholds = errors.New("invalid thresholds")
	ErrInvalidParameters = errors.New("invalid parameters")
)

func (b *evaluationRuleBuilder) Build() (EvaluationRule, error) {
	result := EvaluationRule{
		ID:         ID(utils.GenerateUUID()),
		Parameters: make([]EvaluetionRuleParameter, 0),
		Enabled:    true,
	}
	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return EvaluationRule{}, err
		}
	}

	if !validatorByKind[result.Kind](result.Parameters) {
		return EvaluationRule{}, ErrInvalidParameters
	}

	return result, nil
}

var validatorByKind = map[string]func([]EvaluetionRuleParameter) bool{
	"time": func(params []EvaluetionRuleParameter) bool { return false },
	"threshold": func(params []EvaluetionRuleParameter) bool {
		return utils.AllTrue(
			utils.SomeHasFieldWithValue(params, "Key", "metric"),
			utils.SomeHasFieldWithValue(params, "Key", "lower_threshold"),
			utils.SomeHasFieldWithValue(params, "Key", "upper_threshold"),
		)
	},
}
