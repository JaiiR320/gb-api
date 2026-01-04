# Cache Refactor Design

## Current State

Three separate cache implementations with duplicated boilerplate:

- `track/bigdata/cache.go` - `Cache` for `*BigData` headers
- `track/bigdata/rangecache.go` - `RangeCache[T]` for range-based data
- `api/cache.go` - `BigWigCache` embedding `RangeCache` with fetch logic

Each has its own mutex wrapper around `lru.Cache`, duplicating Get/Set/Len patterns.

## Proposed Design

### Generic Base Cache

```go
// track/bigdata/cache/cache.go
package cache

type Cache[V any] struct {
    cache *lru.Cache[string, V]
    mu    sync.RWMutex
    name  string // for logging
    size  int    // max size for logging
}

func New[V any](name string, size int) (*Cache[V], error)
func (c *Cache[V]) Get(key string) (V, bool)
func (c *Cache[V]) Set(key string, value V) (evicted bool)
func (c *Cache[V]) Len() int
func (c *Cache[V]) Keys() []string
```

### RangeCache as Extension

```go
// track/bigdata/cache/rangecache.go
type RangeData[T any] struct {
    Start int
    End   int
    Data  []T
}

type RangeCache[T any] struct {
    *Cache[[]RangeData[T]]
}

func NewRangeCache[T any](name string, size int) (*RangeCache[T], error)
func (c *RangeCache[T]) FindMissingRanges(start, end int, key string) []Range
```

### Domain Usage

```go
// track/bigdata/bigwig/bigwig.go
var HeaderCache = cache.New[*bigdata.BigData]("bigwig-headers", 25)

// api/cache.go
type BigWigCache struct {
    *cache.RangeCache[bigwig.BigWigData]
}

func (b *BigWigCache) GetCachedWigData(url, chrom string, start, end int) ([]bigwig.BigWigData, error) {
    // domain-specific fetch + merge logic
}
```

## Migration Steps

1. Create `track/bigdata/cache/` package with generic `Cache[V]`
2. Add `RangeCache[T]` that embeds `Cache[[]RangeData[T]]`
3. Move `FindRanges` and `Range` types to new package
4. Update `bigwig.go` and `bigbed.go` to use `cache.New[*bigdata.BigData]`
5. Update `api/cache.go` to use `cache.NewRangeCache`
6. Delete old `cache.go` and `rangecache.go`

## Benefits

- Single place for logging, metrics, size tracking
- Consistent API across all caches
- Easy to add features (TTL, callbacks, stats endpoint)
- Domain code focuses on fetch/merge logic, not cache mechanics
