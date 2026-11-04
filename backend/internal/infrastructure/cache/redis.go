package cache

import (
	"context"
	"time"

	"boilerplate-learner/internal/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(cfg *config.Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Redis{client: client}, nil
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// SetNX sets key to value only if it does not already exist, returning whether
// the write happened. Used as a lightweight distributed lock (idempotency keys).
func (r *Redis) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, ttl).Result()
}

// Incr increments the counter at key and, on the first hit, arms a TTL so the
// window expires. Returns the running count. Backs the fixed-window rate limiter.
func (r *Redis) Incr(ctx context.Context, key string, window time.Duration) (int64, error) {
	pipe := r.client.TxPipeline()
	count := pipe.Incr(ctx, key)
	pipe.ExpireNX(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return count.Val(), nil
}

func (r *Redis) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	return r.client.Set(ctx, "blacklist:"+jti, "1", ttl).Err()
}

func (r *Redis) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	return r.Exists(ctx, "blacklist:"+jti)
}

func (r *Redis) Close() error {
	return r.client.Close()
}
