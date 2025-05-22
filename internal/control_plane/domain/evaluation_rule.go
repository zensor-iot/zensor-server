package domain

import (
	"errors"
	"zensor-server/internal/infra/utils"
)

type EvaluationRule struct {
	ID          ID
	Version     Version
	Description string
	Kind        string
	Parameters  []EvaluationRuleParameter
	Enabled     bool
}

func (er *EvaluationRule) AddParameters(params ...EvaluationRuleParameter) {
	er.Parameters = append(er.Parameters, params...)
}

type EvaluationRuleParameter struct {
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

func (b *evaluationRuleBuilder) WithParameters(value ...EvaluationRuleParameter) *evaluationRuleBuilder {
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
		Parameters: make([]EvaluationRuleParameter, 0),
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

var validatorByKind = map[string]func([]EvaluationRuleParameter) bool{
	"time": func(params []EvaluationRuleParameter) bool {
		return utils.AllTrue(
			utils.SomeHasFieldWithValue(params, "Key", "start"),
			utils.SomeHasFieldWithValue(params, "Key", "task"),
		)
	},
	"threshold": func(params []EvaluationRuleParameter) bool {
		return utils.AllTrue(
			utils.SomeHasFieldWithValue(params, "Key", "metric"),
			utils.SomeHasFieldWithValue(params, "Key", "lower_threshold"),
			utils.SomeHasFieldWithValue(params, "Key", "upper_threshold"),
		)
	},
}
