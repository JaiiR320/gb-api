package bigdata

import (
	"fmt"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

type Range struct {
	Start int
	End   int
}

type RangeData[T any] struct {
	Start int
	End   int
	Data  []T
}

type RangeCache[T any] struct {
	Cache *lru.Cache[string, []RangeData[T]]
	Mu    sync.RWMutex
}

func NewRangeCache[T any](size int) (*RangeCache[T], error) {
	cache, err := lru.New[string, []RangeData[T]](size)
	if err != nil {
		return nil, err
	}
	return &RangeCache[T]{
		Cache: cache,
	}, nil
}

func (c *RangeCache[T]) Add(id string, data []RangeData[T]) (evicted bool) {
	c.Mu.Lock()
	evicted = c.Cache.Add(id, data)
	c.Mu.Unlock()
	if evicted {
		fmt.Println("evicted existing cache")
	}
	return evicted
}

func (c *RangeCache[T]) Get(id string) (data []RangeData[T], hit bool) {
	c.Mu.RLock()
	data, hit = c.Cache.Get(id)
	c.Mu.RUnlock()
	return data, hit
}

func (c *RangeCache[T]) Len() (length int) {
	l := c.Cache.Len()
	fmt.Printf("cache size: %d/25\n", l)
	return l
}

// Find non-overlapping ranges given a list of ranges and a requested range
func FindRanges[T any](start, end int, cached []RangeData[T]) []Range {
	out := []Range{}
	current := start
	for _, cache := range cached {
		if current >= cache.End {
			continue
		}

		if current < cache.Start {
			temp := min(cache.Start, end)
			out = append(out, Range{Start: current, End: temp})
			current = temp
			if end < cache.Start {
				return out
			}
		}

		if current >= cache.Start {
			if end <= cache.End {
				return out
			}
			current = cache.End
		}
	}
	if current != end {
		out = append(out, Range{Start: current, End: end})
	}
	return out
}
