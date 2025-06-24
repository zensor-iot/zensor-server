package replication

import (
	"log/slog"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
)

// Service manages the replication module
type Service struct {
	replicator      *Replicator
	consumerFactory pubsub.ConsumerFactory
	orm             sql.ORM
}

// NewService creates a new replication service
func NewService(
	consumerFactory pubsub.ConsumerFactory,
	orm sql.ORM,
) *Service {
	replicator := NewReplicator(consumerFactory, orm)

	return &Service{
		replicator:      replicator,
		consumerFactory: consumerFactory,
		orm:             orm,
	}
}

// Start begins the replication process
func (s *Service) Start() error {
	slog.Info("starting replication service")

	// Start replication
	err := s.replicator.Start()
	if err != nil {
		slog.Error("failed to start replication", slog.Any("error", err))
		return err
	}

	slog.Info("replication service started successfully")

	return nil
}

// Stop gracefully stops the replication service
func (s *Service) Stop() {
	slog.Info("stopping replication service")
	s.replicator.Stop()
	slog.Info("replication service stopped")
}

// RegisterHandler allows registering topic handlers
func (s *Service) RegisterHandler(handler TopicHandler) error {
	return s.replicator.RegisterHandler(handler)
}
