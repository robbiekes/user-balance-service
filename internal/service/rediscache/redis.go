package rediscache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
)

const expire = 1 * time.Second

type RedisLib struct {
	client *redis.Client
	expire time.Duration
}

func NewRedisLib(client *redis.Client) *RedisLib {
	return &RedisLib{
		client: client,
		expire: expire,
	}
}

func (r *RedisLib) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	r.client.Set(ctx, key, data, r.expire)
	return nil
}

func (r *RedisLib) Get(ctx context.Context, key string) (interface{}, error) {
	var err error
	var value interface{}

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(result), &value)
	if err != nil {
		return nil, err
	}
	return &value, err
}
