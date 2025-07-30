package middleware

import (
	"context"
	"sync"
	"time"

	"go-cqrs/internal/query"
)

type CacheEntry struct {
	Result    interface{}
	ExpiresAt time.Time
}

func (e CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

type Cache interface {
	Get(key string) (interface{}, bool)

	Set(key string, result interface{}, ttl time.Duration)

	Delete(key string)

	Clear()
}

type MemoryCache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
}

func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		entries: make(map[string]CacheEntry),
	}

	go cache.cleanup()

	return cache
}

func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists || entry.IsExpired() {
		return nil, false
	}

	return entry.Result, true
}

func (c *MemoryCache) Set(key string, result interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = CacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(ttl),
	}
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]CacheEntry)
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entries {
			if entry.IsExpired() {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

type CacheKeyGenerator interface {
	GenerateKey(ctx context.Context, q query.Query) string
}

type DefaultCacheKeyGenerator struct{}

func (g *DefaultCacheKeyGenerator) GenerateKey(ctx context.Context, q query.Query) string {

	return q.QueryName()
}

func CachingMiddleware(cache Cache, keyGen CacheKeyGenerator, defaultTTL time.Duration) query.QueryMiddleware {
	if keyGen == nil {
		keyGen = &DefaultCacheKeyGenerator{}
	}

	return func(next query.QueryHandlerFunc) query.QueryHandlerFunc {
		return func(ctx context.Context, q query.Query) (interface{}, error) {

			cacheKey := keyGen.GenerateKey(ctx, q)

			if cached, found := cache.Get(cacheKey); found {
				return cached, nil
			}

			result, err := next(ctx, q)
			if err != nil {
				return result, err
			}

			cache.Set(cacheKey, result, defaultTTL)

			return result, nil
		}
	}
}

func CachingMiddlewareWithTTL(cache Cache, keyGen CacheKeyGenerator, ttlFunc func(query.Query) time.Duration) query.QueryMiddleware {
	if keyGen == nil {
		keyGen = &DefaultCacheKeyGenerator{}
	}

	return func(next query.QueryHandlerFunc) query.QueryHandlerFunc {
		return func(ctx context.Context, q query.Query) (interface{}, error) {

			cacheKey := keyGen.GenerateKey(ctx, q)

			if cached, found := cache.Get(cacheKey); found {
				return cached, nil
			}

			result, err := next(ctx, q)
			if err != nil {
				return result, err
			}

			ttl := ttlFunc(q)
			if ttl > 0 {
				cache.Set(cacheKey, result, ttl)
			}

			return result, nil
		}
	}
}
