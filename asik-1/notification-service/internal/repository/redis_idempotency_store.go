package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisIdempotencyStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisIdempotencyStore(addr string, ttl time.Duration) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
		ttl: ttl,
	}
}

func (s *RedisIdempotencyStore) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	count, err := s.client.Exists(ctx, s.key(eventID)).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *RedisIdempotencyStore) MarkProcessed(ctx context.Context, eventID string) error {
	if err := s.client.Set(ctx, s.key(eventID), "processed", s.ttl).Err(); err != nil {
		return err
	}

	log.Printf("[idempotency] marked processed event_id=%s ttl=%s", eventID, s.ttl)
	return nil
}

func (s *RedisIdempotencyStore) Close() error {
	return s.client.Close()
}

func (s *RedisIdempotencyStore) key(eventID string) string {
	return fmt.Sprintf("notification:event:%s", eventID)
}
