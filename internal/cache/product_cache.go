package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Timoth26/GopherCart-Bridge/internal/domain"
)

type ProductCache struct {
	client *redis.Client
}

func NewProductCache(client *redis.Client) *ProductCache {
	return &ProductCache{client: client}
}

func (c *ProductCache) key(id int64) string {
	return fmt.Sprintf("product:%d", id)
}

func (c *ProductCache) Get(ctx context.Context, id int64) (*domain.Product, error) {
	val, err := c.client.Get(ctx, c.key(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, domain.ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	}

	var p domain.Product
	if err := json.Unmarshal(val, &p); err != nil {
		return nil, fmt.Errorf("cache unmarshal: %w", err)
	}

	return &p, nil
}

func (c *ProductCache) Set(ctx context.Context, p *domain.Product, ttl time.Duration) error {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}

	if err := c.client.Set(ctx, c.key(p.ID), data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}

	return nil
}

func (c *ProductCache) Delete(ctx context.Context, id int64) error {
	if err := c.client.Del(ctx, c.key(id)).Err(); err != nil {
		return fmt.Errorf("cache delete: %w", err)
	}

	return nil
}
