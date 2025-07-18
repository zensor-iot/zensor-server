package persistence

import (
	"context"
	"testing"
	"time"
	"zensor-server/internal/control_plane/persistence/internal"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/sql"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommand represents the device_commands_final table for testing
type TestCommand struct {
	ID            string `gorm:"primaryKey"`
	Version       int
	DeviceName    string
	DeviceID      string
	TaskID        string
	Payload       string `gorm:"type:json"`
	DispatchAfter string
	Port          uint8
	Priority      string
	CreatedAt     string
	Ready         bool
	Sent          bool
	SentAt        string
}

func (t TestCommand) TableName() string {
	return "device_commands"
}

// setupTestTables creates tables for testing that mimic the Materialized View structure
func setupTestTables(orm sql.ORM) error {
	err := orm.AutoMigrate(&TestCommand{})
	return err
}

func TestNewCommandRepository(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations")
	require.NoError(t, err)

	// Create test tables
	err = setupTestTables(orm)
	require.NoError(t, err)

	// Create a mock publisher factory for testing
	mockFactory := pubsub.NewMemoryPublisherFactory()

	repo, err := NewCommandRepository(orm, mockFactory)
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestSimpleCommandRepository_FindAllPending(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations")
	require.NoError(t, err)

	// Create test tables
	err = setupTestTables(orm)
	require.NoError(t, err)

	// Create a mock publisher factory for testing
	mockFactory := pubsub.NewMemoryPublisherFactory()

	repo, err := NewCommandRepository(orm, mockFactory)
	require.NoError(t, err)

	ctx := context.Background()
	commands, err := repo.FindAllPending(ctx)
	require.NoError(t, err)
	assert.Empty(t, commands)
}

func TestSimpleCommandRepository_FindPendingByDevice(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations")
	require.NoError(t, err)

	// Create test tables
	err = setupTestTables(orm)
	require.NoError(t, err)

	// Create a mock publisher factory for testing
	mockFactory := pubsub.NewMemoryPublisherFactory()

	repo, err := NewCommandRepository(orm, mockFactory)
	require.NoError(t, err)

	ctx := context.Background()
	deviceID := domain.ID("test-device-id")
	commands, err := repo.FindPendingByDevice(ctx, deviceID)
	require.NoError(t, err)
	assert.Empty(t, commands)
}

func TestSimpleCommandRepository_FindByTaskID(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations")
	require.NoError(t, err)

	// Create test tables
	err = setupTestTables(orm)
	require.NoError(t, err)

	// Create a mock publisher factory for testing
	mockFactory := pubsub.NewMemoryPublisherFactory()

	repo, err := NewCommandRepository(orm, mockFactory)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := domain.ID("test-task-id")
	commands, err := repo.FindByTaskID(ctx, taskID)
	require.NoError(t, err)
	assert.Empty(t, commands)
}

func TestSimpleCommandRepository_Create(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations")
	require.NoError(t, err)

	// Create test tables
	err = setupTestTables(orm)
	require.NoError(t, err)

	// Create a mock publisher factory for testing
	mockFactory := pubsub.NewMemoryPublisherFactory()

	repo, err := NewCommandRepository(orm, mockFactory)
	require.NoError(t, err)

	ctx := context.Background()
	cmd := domain.Command{
		ID:       domain.ID("test-command-id"),
		Version:  domain.Version(1),
		Device:   domain.Device{ID: domain.ID("test-device-id"), Name: "test-device"},
		Task:     domain.Task{ID: domain.ID("test-task-id-create")},
		Port:     domain.Port(15),
		Priority: domain.CommandPriority("NORMAL"),
		Payload: domain.CommandPayload{
			Index: domain.Index(0),
			Value: domain.CommandValue(100),
		},
		DispatchAfter: utils.Time{Time: time.Now()},
		Ready:         false,
		Sent:          false,
		CreatedAt:     utils.Time{Time: time.Now()},
	}

	err = repo.Create(ctx, cmd)
	require.NoError(t, err)

	// Note: In the event-driven architecture, the database update happens through
	// the replication layer consuming Kafka events, not directly in the repository
}

func TestSimpleCommandRepository_Update(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations")
	require.NoError(t, err)

	// Create test tables
	err = setupTestTables(orm)
	require.NoError(t, err)

	// Create a mock publisher factory for testing
	mockFactory := pubsub.NewMemoryPublisherFactory()

	repo, err := NewCommandRepository(orm, mockFactory)
	require.NoError(t, err)

	ctx := context.Background()
	cmd := domain.Command{
		ID:       domain.ID("test-command-id-update"),
		Version:  domain.Version(1),
		Device:   domain.Device{ID: domain.ID("test-device-id"), Name: "test-device"},
		Task:     domain.Task{ID: domain.ID("test-task-id-update")},
		Port:     domain.Port(15),
		Priority: domain.CommandPriority("NORMAL"),
		Payload: domain.CommandPayload{
			Index: domain.Index(0),
			Value: domain.CommandValue(100),
		},
		DispatchAfter: utils.Time{Time: time.Now()},
		Ready:         false,
		Sent:          false,
		CreatedAt:     utils.Time{Time: time.Now()},
	}

	// First create the command in the database (simulating replication layer)
	internalCmd := internal.FromCommand(cmd)
	err = orm.WithContext(ctx).Create(&internalCmd).Error()
	require.NoError(t, err)

	// Now update it through the repository (which publishes to Kafka)
	cmd.Ready = true
	cmd.Sent = true
	cmd.SentAt = utils.Time{Time: time.Now()}

	err = repo.Update(ctx, cmd)
	require.NoError(t, err)

	// Note: In the event-driven architecture, the database update happens through
	// the replication layer consuming Kafka events, not directly in the repository
}

func TestSimpleCommandRepository_Update_NotFound(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations")
	require.NoError(t, err)

	// Create test tables
	err = setupTestTables(orm)
	require.NoError(t, err)

	// Create a mock publisher factory for testing
	mockFactory := pubsub.NewMemoryPublisherFactory()

	repo, err := NewCommandRepository(orm, mockFactory)
	require.NoError(t, err)

	ctx := context.Background()
	cmd := domain.Command{
		ID:       domain.ID("non-existent-command-id"),
		Version:  domain.Version(1),
		Device:   domain.Device{ID: domain.ID("test-device-id"), Name: "test-device"},
		Task:     domain.Task{ID: domain.ID("test-task-id")},
		Port:     domain.Port(15),
		Priority: domain.CommandPriority("NORMAL"),
		Payload: domain.CommandPayload{
			Index: domain.Index(0),
			Value: domain.CommandValue(100),
		},
		DispatchAfter: utils.Time{Time: time.Now()},
		Ready:         true,
		Sent:          true,
		CreatedAt:     utils.Time{Time: time.Now()},
	}

	// Try to update a non-existent command
	err = repo.Update(ctx, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command not found")
}
