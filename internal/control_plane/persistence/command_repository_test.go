package persistence

import (
	"context"
	"testing"
	"zensor-server/internal/control_plane/domain"
	"zensor-server/internal/infra/sql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommandRepository(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations", nil)
	require.NoError(t, err)

	repo, err := NewCommandRepository(orm)
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestSimpleCommandRepository_FindAllPending(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations", nil)
	require.NoError(t, err)

	repo, err := NewCommandRepository(orm)
	require.NoError(t, err)

	ctx := context.Background()
	commands, err := repo.FindAllPending(ctx)
	require.NoError(t, err)
	assert.Empty(t, commands)
}

func TestSimpleCommandRepository_FindPendingByDevice(t *testing.T) {
	orm, err := sql.NewMemoryORM("migrations", nil)
	require.NoError(t, err)

	repo, err := NewCommandRepository(orm)
	require.NoError(t, err)

	ctx := context.Background()
	deviceID := domain.ID("test-device-id")
	commands, err := repo.FindPendingByDevice(ctx, deviceID)
	require.NoError(t, err)
	assert.Empty(t, commands)
}
