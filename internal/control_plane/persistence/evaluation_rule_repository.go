package persistence

import (
	"context"
	"fmt"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/domain"
)

var _ usecases.EvaluationRuleRepository = (*EvaluationRuleRepository)(nil)

func NewEvaluationRuleRepository(orm sql.ORM) (*EvaluationRuleRepository, error) {
	err := orm.AutoMigrate(&internal.EvaluationRule{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	return &EvaluationRuleRepository{
		orm: orm,
	}, nil
}

type EvaluationRuleRepository struct {
	orm sql.ORM
}

func (e *EvaluationRuleRepository) AddToDevice(ctx context.Context, device domain.Device, evaluationRule domain.EvaluationRule) error {
	entity := internal.FromEvaluationRule(evaluationRule)
	entity.DeviceID = device.ID.String()

	err := e.orm.WithContext(ctx).Create(&entity).Error()
	if err != nil {
		return fmt.Errorf("creating evaluation rule in database: %w", err)
	}

	return nil
}

func (e *EvaluationRuleRepository) FindAllByDeviceID(ctx context.Context, deviceID string) ([]domain.EvaluationRule, error) {
	var rules []internal.EvaluationRule
	err := e.
		orm.
		WithContext(ctx).
		Where("device_id = ?", deviceID).
		Find(&rules).
		Error()
	if err != nil {
		return nil, fmt.Errorf("query evaluation rules: %w", err)
	}

	domainRules := make([]domain.EvaluationRule, len(rules))
	for i, r := range rules {
		domainRules[i] = r.ToDomain()
	}

	return domainRules, nil
}
