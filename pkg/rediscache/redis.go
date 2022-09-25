package rediscache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
)

const defaultExpire = 300 * time.Second

type Redis struct {
	client *redis.Client
	expire time.Duration
}

func New(client *redis.Client, opts ...Option) *Redis {
	r := &Redis{
		client: client,
		expire: defaultExpire,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *Redis) Set(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	r.client.Set(ctx, key, data, r.expire)
	return nil
}

func (r *Redis) Get(ctx context.Context, key string) (any, error) {
	var value any

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(result), &value)
	if err != nil {
		return nil, err
	}

	return value, nil
}
