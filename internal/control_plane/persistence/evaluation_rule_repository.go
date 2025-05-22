package persistence

import (
	"context"
	"fmt"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
)

var _ usecases.EvaluationRuleRepository = (*EvaluationRuleRepository)(nil)

func NewEvaluationRuleRepository(publisherFactory pubsub.PublisherFactory, orm sql.ORM) (*EvaluationRuleRepository, error) {
	publisher, err := publisherFactory.New("evaluation_rules", internal.EvaluationRule{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	return &EvaluationRuleRepository{
		publisher: publisher,
		orm:       orm,
	}, nil
}

type EvaluationRuleRepository struct {
	publisher pubsub.Publisher
	orm       sql.ORM
}

func (e *EvaluationRuleRepository) AddToDevice(ctx context.Context, device domain.Device, evaluationRule domain.EvaluationRule) error {
	data := internal.FromEvaluationRule(evaluationRule)
	data.DeviceID = string(device.ID)
	err := e.publisher.Publish(ctx, pubsub.Key(evaluationRule.ID), data)
	if err != nil {
		return fmt.Errorf("publishing to kafka: %w", err)
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
