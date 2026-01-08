package cache

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

type Cache[T any] struct {
	Cache *lru.Cache[string, T]
	Mu    sync.RWMutex
}

func NewCache[T any](size int) (*Cache[T], error) {
	cache, err := lru.New[string, T](size)
	if err != nil {
		return nil, err
	}
	return &Cache[T]{
		Cache: cache,
	}, nil
}

func (c *Cache[T]) Add(key string, val T) (evicted bool) {
	c.Mu.Lock()
	evicted = c.Cache.Add(key, val)
	c.Mu.Unlock()
	return evicted
}

func (c *Cache[T]) Get(key string) (val T, hit bool) {
	c.Mu.RLock()
	val, hit = c.Cache.Get(key)
	c.Mu.RUnlock()
	return val, hit
}

func (c *Cache[T]) Len() (length int) {
	return c.Cache.Len()
}

func (c *Cache[T]) Keys() []string {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	return c.Cache.Keys()
}
