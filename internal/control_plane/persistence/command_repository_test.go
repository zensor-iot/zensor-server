package persistence

import (
	"context"
	"testing"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/infra/sql"

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

	repo, err := NewCommandRepository(orm)
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestSimpleCommandRepository_FindAllPending(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations")
	require.NoError(t, err)

	// Create test tables
	err = setupTestTables(orm)
	require.NoError(t, err)

	repo, err := NewCommandRepository(orm)
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

	repo, err := NewCommandRepository(orm)
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

	repo, err := NewCommandRepository(orm)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := domain.ID("test-task-id")
	commands, err := repo.FindByTaskID(ctx, taskID)
	require.NoError(t, err)
	assert.Empty(t, commands)
}
