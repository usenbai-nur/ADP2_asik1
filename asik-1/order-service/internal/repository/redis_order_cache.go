package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ap2/order-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type RedisOrderCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisOrderCache(addr string, ttl time.Duration) *RedisOrderCache {
	return &RedisOrderCache{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
		ttl: ttl,
	}
}

func (c *RedisOrderCache) Get(ctx context.Context, id string) (*domain.Order, error) {
	data, err := c.client.Get(ctx, c.key(id)).Result()
	if errors.Is(err, redis.Nil) {
		log.Printf("[order-cache] MISS order_id=%s", id)
		return nil, ErrCacheMiss
	}
	if err != nil {
		log.Printf("[order-cache] ERROR get order_id=%s error=%v", id, err)
		return nil, err
	}

	var order domain.Order
	if err := json.Unmarshal([]byte(data), &order); err != nil {
		return nil, err
	}

	log.Printf("[order-cache] HIT order_id=%s", id)
	return &order, nil
}

func (c *RedisOrderCache) Set(ctx context.Context, order *domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	if err := c.client.Set(ctx, c.key(order.ID), data, c.ttl).Err(); err != nil {
		log.Printf("[order-cache] ERROR set order_id=%s error=%v", order.ID, err)
		return err
	}

	log.Printf("[order-cache] SET order_id=%s ttl=%s", order.ID, c.ttl)
	return nil
}

func (c *RedisOrderCache) Delete(ctx context.Context, id string) error {
	if err := c.client.Del(ctx, c.key(id)).Err(); err != nil {
		log.Printf("[order-cache] ERROR delete order_id=%s error=%v", id, err)
		return err
	}

	log.Printf("[order-cache] INVALIDATED order_id=%s", id)
	return nil
}

func (c *RedisOrderCache) Close() error {
	return c.client.Close()
}

func (c *RedisOrderCache) key(id string) string {
	return fmt.Sprintf("order:%s", id)
}