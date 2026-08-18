package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"zensor-server/internal/shared_kernel/domain"
	"zensor-server/internal/shared_kernel/usecases"

	"github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix        = "auth:session:"
	sessionsByUserKeyPrefix = "auth:sessions_by_user:"
)

func NewRedisSessionStore(client redis.Cmdable) *RedisSessionStore {
	return &RedisSessionStore{
		client: client,
	}
}

var _ usecases.SessionStore = (*RedisSessionStore)(nil)

// RedisSessionStore persists sessions in Redis, indexed per user for instant revocation.
type RedisSessionStore struct {
	client redis.Cmdable
}

func (s *RedisSessionStore) Create(ctx context.Context, session domain.Session) error {
	payload, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return errors.New("session already expired")
	}

	if err := s.client.Set(ctx, sessionKeyPrefix+session.ID, payload, ttl).Err(); err != nil {
		return fmt.Errorf("storing session: %w", err)
	}

	userKey := sessionsByUserKeyPrefix + session.UserID.String()
	if err := s.client.SAdd(ctx, userKey, session.ID).Err(); err != nil {
		return fmt.Errorf("indexing session by user: %w", err)
	}
	if err := s.client.ExpireGT(ctx, userKey, ttl).Err(); err != nil {
		return fmt.Errorf("setting user index expiry: %w", err)
	}

	return nil
}

func (s *RedisSessionStore) Get(ctx context.Context, sessionID string) (domain.Session, error) {
	payload, err := s.client.Get(ctx, sessionKeyPrefix+sessionID).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.Session{}, usecases.ErrSessionNotFound
	}
	if err != nil {
		return domain.Session{}, fmt.Errorf("fetching session: %w", err)
	}

	var session domain.Session
	if err := json.Unmarshal(payload, &session); err != nil {
		return domain.Session{}, fmt.Errorf("unmarshaling session: %w", err)
	}

	if session.IsExpired(time.Now()) {
		return domain.Session{}, usecases.ErrSessionNotFound
	}

	return session, nil
}

func (s *RedisSessionStore) Delete(ctx context.Context, sessionID string) error {
	session, err := s.Get(ctx, sessionID)
	if errors.Is(err, usecases.ErrSessionNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	if err := s.client.Del(ctx, sessionKeyPrefix+sessionID).Err(); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	userKey := sessionsByUserKeyPrefix + session.UserID.String()
	if err := s.client.SRem(ctx, userKey, sessionID).Err(); err != nil {
		return fmt.Errorf("removing session from user index: %w", err)
	}

	return nil
}

func (s *RedisSessionStore) DeleteByUser(ctx context.Context, userID domain.ID) error {
	userKey := sessionsByUserKeyPrefix + userID.String()
	sessionIDs, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("listing user sessions: %w", err)
	}

	for _, sessionID := range sessionIDs {
		if err := s.client.Del(ctx, sessionKeyPrefix+sessionID).Err(); err != nil {
			return fmt.Errorf("deleting session: %w", err)
		}
	}

	if err := s.client.Del(ctx, userKey).Err(); err != nil {
		return fmt.Errorf("deleting user session index: %w", err)
	}

	return nil
}
