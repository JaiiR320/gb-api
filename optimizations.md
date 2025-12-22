# Caching Strategy and Optimization Recommendations

## Project Overview

**Type:** Genomics Data API (REST API)
**Tech Stack:** Go 1.25, HTTP server (stdlib), Docker/Fly.io deployment
**Primary Purpose:** Serves genomic track data for a genome browser application

**Key Files:**
- `main.go` - Main server entry point
- `api/handlers.go` - Request handlers
- `track/bigdata/requestbytes.go` - HTTP range requests
- `track/transcript/records.go` - Local GTF file access

---

## Data Access Patterns Identified

### A. Remote BigWig/BigBed Files (External HTTP)
- **Location:** Remote URLs (e.g., `downloads.wenglab.org`)
- **Access Pattern:** HTTP Range requests for specific genomic regions
- **Operations:**
  - Header loading (64 bytes at file start)
  - Metadata loading (zoom levels, chromosome tree, etc.)
  - R+ tree traversal (multiple small range requests)
  - Data block retrieval (variable size, compressed with zlib)
- **Frequency:** Every request makes 5-10+ HTTP calls per track
- **Bottleneck:** Network latency for multiple sequential HTTP requests

### B. Local Transcript Data (File I/O)
- **Location:** `./track/transcript/data/v40/sorted.gtf.gz`
- **Access Pattern:** Indexed tabix queries using bix library
- **Operations:**
  - File open/close on every request
  - Index lookup via `.tbi` file
  - Region-specific decompression
- **Frequency:** Every transcript request
- **Bottleneck:** File I/O and decompression overhead

### C. Browser Endpoint (Concurrent Requests)
- **Location:** `/browser` endpoint in `api/handlers.go:56`
- **Pattern:** Spawns goroutines for parallel track fetching
- **Operations:** Multiple tracks fetched concurrently using channels
- **Bottleneck:** Each goroutine still does full data fetch without caching

---

## Current Caching Status

**NO CACHING EXISTS**

**Current Issues:**
1. New HTTP client created for every `RequestBytes()` call
2. BigWig/BigBed metadata (headers, trees) re-fetched on every request
3. GTF file opened/closed on every transcript request
4. No connection pooling or HTTP client reuse
5. No in-memory caching of frequently accessed data

---

## Performance Bottlenecks

### Critical Issues:
1. **Repeated metadata fetching** - Same BigWig/BigBed files have headers/trees loaded multiple times
2. **No HTTP client reuse** - Creating `&http.Client{}` on every request in `track/bigdata/requestbytes.go:18`
3. **File handle churn** - GTF file opened/closed repeatedly
4. **No decompression caching** - Zlib decompression repeated for overlapping regions
5. **Redundant tree traversals** - R+ tree navigation repeated for same regions

---

## Recommended Caching Strategy

### TIER 1: IMMEDIATE HIGH-IMPACT CHANGES

#### A. HTTP Client Connection Pool (CRITICAL)
**Location:** `track/bigdata/requestbytes.go`

**Implementation:**
```go
var httpClient = &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 20,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  true, // Handle compression manually
    },
}
```

**Impact:** Reduces connection overhead, enables HTTP keep-alive
**Effort:** 5 minutes
**Performance Gain:** 20-30% reduction in request latency

---

#### B. BigData Metadata Cache (HIGH PRIORITY)
**Location:** Create new file `track/bigdata/cache.go`

**Strategy:** In-memory LRU cache for BigWig/BigBed metadata (headers, trees, zoom levels)

**Implementation Approach:**
```go
type MetadataCache struct {
    mu    sync.RWMutex
    cache *lru.Cache // github.com/hashicorp/golang-lru
}

type CachedMetadata struct {
    Header       Header
    ZoomLevels   []ZoomLevelHeader
    ByteOrder    binary.ByteOrder
    AutoSql      string
    TotalSummary TotalSummary
    ChromTree    ChromTree
    CachedAt     time.Time
}
```

**Cache Key:** URL hash
**TTL:** 1 hour (metadata rarely changes)
**Max Size:** 100 entries (~50MB memory)

**Impact:** Eliminates 4-6 HTTP requests per BigWig/BigBed query
**Effort:** 2-3 hours
**Performance Gain:** 40-60% faster BigWig/BigBed queries

---

#### C. GTF File Handle Pooling
**Location:** `track/transcript/records.go`

**Current Issue:** Opens/closes file on every request (lines 66-67)

**Strategy:** Keep tabix file handle open or use a pool

**Implementation:**
```go
var (
    gtfMutex sync.Mutex
    gtfTbx   *bix.Bix
)

func GetRecords(pathStr string, posStr string) ([]Record, error) {
    gtfMutex.Lock()
    if gtfTbx == nil {
        tbx, err := bix.New(pathStr)
        if err != nil {
            gtfMutex.Unlock()
            return nil, err
        }
        gtfTbx = tbx
    }
    gtfMutex.Unlock()
    // Use gtfTbx for queries...
}
```

**Impact:** Eliminates file open overhead on every transcript request
**Effort:** 1 hour
**Performance Gain:** 30-40% faster transcript queries

---

### TIER 2: MEDIUM-TERM OPTIMIZATIONS

#### D. Regional Data Cache (Medium Priority)
**Purpose:** Cache actual genomic data for frequently accessed regions

**Strategy:** LRU cache with regional keys

**Cache Key Structure:**
```go
type RegionKey struct {
    URL   string
    Chrom string
    Start int32
    End   int32
}
```

**Size:** 500MB - 1GB memory limit
**TTL:** 15-30 minutes
**Eviction:** LRU with size-based limits

**Best For:**
- Popular genomic regions (e.g., frequently viewed genes)
- Browser sessions with repeated panning/zooming
- Shared data files accessed by multiple users

**Implementation Library:** 
- `github.com/hashicorp/golang-lru` (simple)
- `github.com/dgraph-io/ristretto` (more sophisticated)
- `github.com/allegro/bigcache` (high performance, limited features)

**Impact:** 80-90% cache hit rate for repeated regions
**Effort:** 4-6 hours
**Performance Gain:** Near-instant responses for cached regions

---

#### E. R+ Tree Node Cache
**Location:** `track/bigdata/rtree.go`

**Strategy:** Cache R+ tree nodes to avoid repeated traversals

**Implementation:**
```go
type TreeNodeCache struct {
    cache map[string][]RPLeafNode // URL+region -> leaf nodes
    mu    sync.RWMutex
}
```

**Impact:** Reduces HTTP requests for tree navigation
**Effort:** 2-3 hours
**Performance Gain:** 25-35% fewer HTTP requests

---

### TIER 3: ADVANCED/FUTURE OPTIMIZATIONS

#### F. Redis Distributed Cache
**When:** If scaling to multiple server instances

**Use Cases:**
- Share metadata cache across instances
- Persist popular region data
- Handle cache warming strategies

**Trade-offs:**
- Adds infrastructure complexity
- Network overhead for cache lookups
- Better for high-traffic, multi-instance deployments

---

#### G. CDN/Reverse Proxy Cache
**Options:** Cloudflare, Fastly, Varnish

**Configuration:** Cache based on POST body content hash
**Benefit:** Offload repeated identical queries
**Limitation:** Requires predictable query patterns

---

#### H. Prefetching/Read-ahead
**Strategy:** Predict likely next regions based on browser navigation

**Implementation:**
- When user views chr1:1000-2000, prefetch chr1:2000-3000
- Use goroutines for background prefetching
- Track user navigation patterns

**Complexity:** High
**Benefit:** Reduces perceived latency

---

## Recommended Implementation Order

### Phase 1 (Week 1) - Quick Wins:
1. HTTP client connection pooling
2. BigData metadata caching
3. GTF file handle pooling

**Expected Result:** 50-60% performance improvement with minimal code changes

### Phase 2 (Week 2-3) - Data Caching:
4. Regional data cache with LRU eviction
5. R+ tree node caching
6. Cache monitoring/metrics

**Expected Result:** 70-80% improvement for repeated queries

### Phase 3 (Month 2+) - Advanced:
7. Consider Redis if scaling horizontally
8. Evaluate CDN options for static-like queries
9. Implement intelligent prefetching

---

## Specific Code Locations to Modify

1. `track/bigdata/requestbytes.go:18` - Add global HTTP client
2. `track/bigdata/loader.go:13,54` - Add metadata cache lookups
3. `track/transcript/records.go:66-67` - Add file handle pooling
4. `track/bigdata/reader.go:25,50` - Add data caching
5. `api/handlers.go:78` - Add cache warming in concurrent fetches

---

## Memory Budget Recommendations

**Conservative Deployment (512MB container):**
- Metadata cache: 50MB (100 files)
- Regional data cache: 200MB
- Application overhead: 262MB

**Recommended Deployment (1GB container):**
- Metadata cache: 100MB (200 files)
- Regional data cache: 600MB
- Application overhead: 324MB

**High-performance Deployment (2GB container):**
- Metadata cache: 200MB (400 files)
- Regional data cache: 1.5GB
- Application overhead: 324MB

---

## Monitoring Recommendations

Add cache metrics:
- Cache hit/miss rates
- Cache eviction frequency
- HTTP request counts per query
- Average response times
- Memory usage by cache type

Use Go's `expvar` package or integrate Prometheus metrics.

---

## Summary

**Best Strategy for This Project:** Multi-layer in-memory caching with LRU eviction

**Why:**
1. High read-to-write ratio (no data modifications)
2. Expensive network operations (multiple HTTP range requests)
3. Predictable access patterns (genomic regions)
4. Limited data variety (same files accessed repeatedly)
5. Small metadata footprint (headers/trees are tiny vs data)

**Start with Phase 1** - you'll see immediate, dramatic improvements with minimal risk and effort.
