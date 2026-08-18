package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"zensor-server/internal/shared_kernel/domain"
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
	rules, err := s.evaluationRuleRepository.FindAllByDeviceID(ctx, string(device.ID))
	if err != nil {
		return nil, fmt.Errorf("finding all rules by device id: %w", err)
	}

	return rules, nil
}
