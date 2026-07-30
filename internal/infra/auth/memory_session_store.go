package auth

import (
	"context"
	"sync"
	"time"
	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/usecases"
)

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions:       make(map[string]domain.Session),
		sessionsByUser: make(map[domain.ID]map[string]struct{}),
	}
}

var _ usecases.SessionStore = (*MemorySessionStore)(nil)

// MemorySessionStore keeps sessions in process memory for ENV=local runs and tests.
type MemorySessionStore struct {
	mu             sync.RWMutex
	sessions       map[string]domain.Session
	sessionsByUser map[domain.ID]map[string]struct{}
}

func (s *MemorySessionStore) Create(_ context.Context, session domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	if s.sessionsByUser[session.UserID] == nil {
		s.sessionsByUser[session.UserID] = make(map[string]struct{})
	}
	s.sessionsByUser[session.UserID][session.ID] = struct{}{}

	return nil
}

func (s *MemorySessionStore) Get(_ context.Context, sessionID string) (domain.Session, error) {
	s.mu.RLock()
	session, found := s.sessions[sessionID]
	s.mu.RUnlock()

	if !found {
		return domain.Session{}, usecases.ErrSessionNotFound
	}

	if session.IsExpired(time.Now()) {
		s.mu.Lock()
		s.removeLocked(session)
		s.mu.Unlock()
		return domain.Session{}, usecases.ErrSessionNotFound
	}

	return session, nil
}

func (s *MemorySessionStore) Delete(_ context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, found := s.sessions[sessionID]; found {
		s.removeLocked(session)
	}

	return nil
}

func (s *MemorySessionStore) DeleteByUser(_ context.Context, userID domain.ID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for sessionID := range s.sessionsByUser[userID] {
		delete(s.sessions, sessionID)
	}
	delete(s.sessionsByUser, userID)

	return nil
}

func (s *MemorySessionStore) removeLocked(session domain.Session) {
	delete(s.sessions, session.ID)
	if ids := s.sessionsByUser[session.UserID]; ids != nil {
		delete(ids, session.ID)
		if len(ids) == 0 {
			delete(s.sessionsByUser, session.UserID)
		}
	}
}
