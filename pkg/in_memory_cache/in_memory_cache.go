package in_memory_cache

import (
	"fmt"
	"sync"
	"time"
)

type CacheItem[T any] struct {
	ExpiresAt int64
	Value     T
}

type Cache[T any] struct {
	ttl   time.Duration
	mu    sync.RWMutex
	items map[string]CacheItem[T]
}

func NewMemoryCache[T any](ttl time.Duration) *Cache[T] {
	cache := &Cache[T]{
		ttl:   ttl,
		mu:    sync.RWMutex{},
		items: make(map[string]CacheItem[T]),
	}

	cache.startEvictionLoop()

	return cache
}

func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CacheItem[T]{
		ExpiresAt: time.Now().Add(c.ttl).UnixNano(),
		Value:     value,
	}
}

func (c *Cache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *Cache[T]) Get(key string) (*T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cache_item, exists := c.items[key]
	if !exists {
		return nil, fmt.Errorf("the item with this key does not exist")
	} else {
		return &cache_item.Value, nil
	}
}

func (c *Cache[T]) startEvictionLoop() {
	ticker := time.NewTicker(time.Minute)

	for range ticker.C {
		c.evictExpired()
	}
}

func (c *Cache[T]) evictExpired() {
	c.mu.RLock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	for key, item := range c.items {
		if item.ExpiresAt <= now {
			delete(c.items, key)
		}
	}
}
