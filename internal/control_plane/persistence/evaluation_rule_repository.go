package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/shared_kernel/avro"
)

var _ usecases.EvaluationRuleRepository = (*EvaluationRuleRepository)(nil)

func NewEvaluationRuleRepository(publisherFactory pubsub.PublisherFactory, orm sql.ORM) (*EvaluationRuleRepository, error) {
	publisher, err := publisherFactory.New("evaluation_rules", &avro.AvroEvaluationRule{})
	if err != nil {
		return nil, fmt.Errorf("creating publisher: %w", err)
	}

	err = orm.AutoMigrate(&internal.EvaluationRule{})
	if err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
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
	// Convert domain evaluation rule to Avro evaluation rule
	avroEvaluationRule := &avro.AvroEvaluationRule{
		ID:          evaluationRule.ID.String(),
		DeviceID:    device.ID.String(),
		Version:     int(evaluationRule.Version),
		Description: evaluationRule.Description,
		Kind:        evaluationRule.Kind,
		Enabled:     evaluationRule.Enabled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Convert parameters to JSON string
	parametersJSON, _ := json.Marshal(evaluationRule.Parameters)
	avroEvaluationRule.Parameters = string(parametersJSON)

	err := e.publisher.Publish(ctx, pubsub.Key(evaluationRule.ID), avroEvaluationRule)
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
