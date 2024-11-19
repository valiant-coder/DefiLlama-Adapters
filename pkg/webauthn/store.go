package webauthn

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionStore interface {
	Store(ctx context.Context, key string, data []byte, expiry time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

func NewSessionStore(storeType string) (SessionStore, error) {
	switch storeType {
	case "memory":
		return NewMemorySessionStore(), nil
	case "redis":
		return NewRedisSessionStore(), nil
	}
	return nil, fmt.Errorf("invalid session store type: %s", storeType)
}

type MemorySessionStore struct {
	mu   sync.Mutex
	data map[string][]byte
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		data: make(map[string][]byte),
	}
}

func (s *MemorySessionStore) Store(ctx context.Context, key string, data []byte, expiry time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = data
	return nil
}

func (s *MemorySessionStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[key], nil
}

func (s *MemorySessionStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

type RedisSessionStore struct {
	redis *redis.Client
}

func NewRedisSessionStore() *RedisSessionStore {
	return &RedisSessionStore{}
}

func (s *RedisSessionStore) Store(ctx context.Context, key string, data []byte, expiry time.Duration) error {
	return s.redis.Set(ctx, key, data, expiry).Err()
}

func (s *RedisSessionStore) Get(ctx context.Context, key string) ([]byte, error) {
	return s.redis.Get(ctx, key).Bytes()
}

func (s *RedisSessionStore) Delete(ctx context.Context, key string) error {
	return s.redis.Del(ctx, key).Err()
}
