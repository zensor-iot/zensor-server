package usecases

import (
	"context"
	"errors"
	"zensor-server/internal/control_plane/domain"
)

func NewEvaluationRuleService() *SimpleEvaluationRuleService {
	return &SimpleEvaluationRuleService{}
}

var _ EvaluationRuleService = (*SimpleEvaluationRuleService)(nil)

type SimpleEvaluationRuleService struct {
}

func (s *SimpleEvaluationRuleService) Create(ctx context.Context, evaluationRule domain.EvaluationRule) error {
	return errors.New("not implemented")
}

func (s *SimpleEvaluationRuleService) Update(ctx context.Context, evaluationRule domain.EvaluationRule) error {
	return errors.New("not implemented")
}

func (s *SimpleEvaluationRuleService) Delete(ctx context.Context, id domain.ID) error {
	return errors.New("not implemented")
}

func (s *SimpleEvaluationRuleService) Get(ctx context.Context, id domain.ID) (domain.EvaluationRule, error) {
	return domain.EvaluationRule{}, errors.New("not implemented")
}
