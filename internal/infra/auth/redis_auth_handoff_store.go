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

const authHandoffKeyPrefix = "auth:handoff:"

func NewRedisAuthHandoffStore(client redis.Cmdable) *RedisAuthHandoffStore {
	return &RedisAuthHandoffStore{
		client: client,
	}
}

var _ usecases.AuthHandoffStore = (*RedisAuthHandoffStore)(nil)

// RedisAuthHandoffStore keeps mobile login handoffs in Redis until they expire
// or are consumed.
type RedisAuthHandoffStore struct {
	client redis.Cmdable
}

func (s *RedisAuthHandoffStore) Save(ctx context.Context, codeHash string, handoff domain.AuthHandoff) error {
	payload, err := json.Marshal(handoff)
	if err != nil {
		return fmt.Errorf("marshaling auth handoff: %w", err)
	}

	ttl := time.Until(handoff.ExpiresAt)
	if ttl <= 0 {
		return ErrSessionAlreadyExpired
	}

	if err := s.client.Set(ctx, authHandoffKeyPrefix+codeHash, payload, ttl).Err(); err != nil {
		return fmt.Errorf("storing auth handoff: %w", err)
	}

	return nil
}

// Consume reads and deletes the handoff in one GETDEL, so two concurrent
// exchanges of the same code cannot both succeed.
func (s *RedisAuthHandoffStore) Consume(ctx context.Context, codeHash string) (domain.AuthHandoff, error) {
	payload, err := s.client.GetDel(ctx, authHandoffKeyPrefix+codeHash).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.AuthHandoff{}, usecases.ErrAuthHandoffNotFound
	}
	if err != nil {
		return domain.AuthHandoff{}, fmt.Errorf("consuming auth handoff: %w", err)
	}

	var handoff domain.AuthHandoff
	if err := json.Unmarshal(payload, &handoff); err != nil {
		return domain.AuthHandoff{}, fmt.Errorf("unmarshaling auth handoff: %w", err)
	}

	return handoff, nil
}
