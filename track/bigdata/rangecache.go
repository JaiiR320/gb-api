package bigdata

import (
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
