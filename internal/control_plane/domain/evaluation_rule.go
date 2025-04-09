package domain

import (
	"errors"
	"zensor-server/internal/infra/utils"
)

type EvaluationRule struct {
	ID             ID
	Description    string
	Metric         string
	LowerThreshold float64
	UpperThreshold float64
	Enabled        bool
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

func (b *evaluationRuleBuilder) WithMetric(value string) *evaluationRuleBuilder {
	b.actions = append(b.actions, func(d *EvaluationRule) error {
		d.Metric = value
		return nil
	})
	return b
}

func (b *evaluationRuleBuilder) WithLowerThreshold(value float64) *evaluationRuleBuilder {
	b.actions = append(b.actions, func(d *EvaluationRule) error {
		d.LowerThreshold = value
		return nil
	})
	return b
}

func (b *evaluationRuleBuilder) WithUpperThreshold(value float64) *evaluationRuleBuilder {
	b.actions = append(b.actions, func(d *EvaluationRule) error {
		d.UpperThreshold = value
		return nil
	})
	return b
}

var (
	ErrInvalidThresholds = errors.New("invalid thresholds")
)

func (b *evaluationRuleBuilder) Build() (EvaluationRule, error) {
	result := EvaluationRule{
		ID:      ID(utils.GenerateUUID()),
		Enabled: true,
	}
	for _, a := range b.actions {
		if err := a(&result); err != nil {
			return EvaluationRule{}, err
		}
	}

	if result.LowerThreshold >= result.UpperThreshold {
		return EvaluationRule{}, ErrInvalidThresholds
	}

	return result, nil
}
