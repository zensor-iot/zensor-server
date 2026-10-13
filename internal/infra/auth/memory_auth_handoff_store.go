package auth

import (
	"context"
	"sync"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/usecases"
)

func NewMemoryAuthHandoffStore() *MemoryAuthHandoffStore {
	return &MemoryAuthHandoffStore{
		handoffs: make(map[string]domain.AuthHandoff),
	}
}

var _ usecases.AuthHandoffStore = (*MemoryAuthHandoffStore)(nil)

// MemoryAuthHandoffStore keeps mobile login handoffs in process memory for
// ENV=local runs and tests.
type MemoryAuthHandoffStore struct {
	mu       sync.Mutex
	handoffs map[string]domain.AuthHandoff
}

func (s *MemoryAuthHandoffStore) Save(_ context.Context, codeHash string, handoff domain.AuthHandoff) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handoffs[codeHash] = handoff

	return nil
}

func (s *MemoryAuthHandoffStore) Consume(_ context.Context, codeHash string) (domain.AuthHandoff, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	handoff, found := s.handoffs[codeHash]
	if !found {
		return domain.AuthHandoff{}, usecases.ErrAuthHandoffNotFound
	}
	delete(s.handoffs, codeHash)

	return handoff, nil
}
