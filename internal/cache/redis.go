package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client       *redis.Client
	prefix       string
	maxItemBytes int
}

func NewRedisStore(opts RedisOptions, maxItemBytes int) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         opts.Addr,
		Password:     opts.Password,
		DB:           opts.DB,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	prefix := opts.KeyPrefix
	if prefix == "" {
		prefix = "rsshub_gateway"
	}
	return &RedisStore{client: client, prefix: prefix, maxItemBytes: maxItemBytes}, nil
}

func (r *RedisStore) Provider() string {
	return "redis"
}

func (r *RedisStore) GetResponse(key string) (Entry, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	value, err := r.client.Get(ctx, r.cacheKey(key)).Bytes()
	if err == redis.Nil {
		return Entry{}, false
	}
	if err != nil {
		return Entry{}, false
	}
	var entry Entry
	if err := json.Unmarshal(value, &entry); err != nil {
		return Entry{}, false
	}
	return entry, true
}

func (r *RedisStore) SetResponse(key string, entry Entry, ttl time.Duration) error {
	if r.maxItemBytes > 0 && EntrySize(entry) > r.maxItemBytes {
		return ErrEntryTooLarge
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("encode cache entry: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Set(ctx, r.cacheKey(key), payload, ttl).Err()
}

func (r *RedisStore) GetString(key string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	value, err := r.client.Get(ctx, r.stringKey(key)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (r *RedisStore) SetString(key string, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Set(ctx, r.stringKey(key), value, ttl).Err()
}

func (r *RedisStore) Close() error {
	return r.client.Close()
}

func (r *RedisStore) cacheKey(key string) string {
	return fmt.Sprintf("%s:cache:%s", r.prefix, key)
}

func (r *RedisStore) stringKey(key string) string {
	return fmt.Sprintf("%s:kv:%s", r.prefix, key)
}
