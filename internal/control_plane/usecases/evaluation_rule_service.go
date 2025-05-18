package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"zensor-server/internal/control_plane/domain"
)

func NewEvaluationRuleService(
	evaluationRuleRepository EvaluationRuleRepository,
) *SimpleEvaluationRuleService {
	return &SimpleEvaluationRuleService{
		evaluationRuleRepository: evaluationRuleRepository,
	}
}

var _ EvaluationRuleService = (*SimpleEvaluationRuleService)(nil)

type SimpleEvaluationRuleService struct {
	evaluationRuleRepository EvaluationRuleRepository
}

func (s *SimpleEvaluationRuleService) AddToDevice(ctx context.Context, device domain.Device, evaluationRule domain.EvaluationRule) error {
	err := s.evaluationRuleRepository.AddToDevice(ctx, device, evaluationRule)
	if err != nil {
		slog.Error("adding evaluation rule to device failed", slog.String("error", err.Error()))
		return fmt.Errorf("adding evaluation rule to device: %w", err)
	}

	return nil
}

func (s *SimpleEvaluationRuleService) FindAllByDevice(ctx context.Context, device domain.Device) ([]domain.EvaluationRule, error) {
	return nil, errors.New("not implemented")
}
