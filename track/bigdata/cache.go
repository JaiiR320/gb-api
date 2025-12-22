package bigdata

import (
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

type Cache struct {
	Cache *lru.Cache[string, *BigData]
	Lock  sync.RWMutex
}

func NewCache() *Cache {
	cache, err := lru.New[string, *BigData](25)
	if err != nil {
		panic(err)
	}
	return &Cache{Cache: cache}
}

func (c *Cache) Get(url string) (*BigData, bool) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	return c.Cache.Get(url)
}

func (c *Cache) Set(url string, b *BigData) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.Cache.Add(url, b)
}
