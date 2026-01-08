package api

import (
	"encoding/json"
	"gb-api/cache"
	"gb-api/track/bigdata"
	"gb-api/track/bigdata/bigbed"
	"gb-api/track/bigdata/bigwig"
	"net/http"
	"runtime"
	"unsafe"
)

type CacheStats struct {
	Name         string   `json:"name"`
	EntryCount   int      `json:"entryCount"`
	ApproxSizeKB int64    `json:"approxSizeKB"`
	ApproxSizeMB int64    `json:"approxSizeMB"`
	Keys         []string `json:"keys,omitempty"`
}

type CacheStatusResponse struct {
	Caches        []CacheStats `json:"caches"`
	TotalSizeKB   int64        `json:"totalSizeKB"`
	TotalSizeMB   int64        `json:"totalSizeMB"`
	RuntimeMemory *MemoryStats `json:"runtimeMemory,omitempty"`
}

type MemoryStats struct {
	AllocMB      uint64 `json:"allocMB"`      // Currently allocated heap objects
	TotalAllocMB uint64 `json:"totalAllocMB"` // Cumulative bytes allocated
	SysMB        uint64 `json:"sysMB"`        // Total memory from OS
	NumGC        uint32 `json:"numGC"`        // Number of GC runs
	HeapAllocMB  uint64 `json:"heapAllocMB"`  // Heap allocated
	HeapSysMB    uint64 `json:"heapSysMB"`    // Heap from OS
	HeapIdleMB   uint64 `json:"heapIdleMB"`   // Heap idle
	HeapInUseMB  uint64 `json:"heapInUseMB"`  // Heap in use
	StackInUseMB uint64 `json:"stackInUseMB"` // Stack in use
}

func CacheSizeHandler(w http.ResponseWriter, r *http.Request) {
	includeKeys := r.URL.Query().Get("keys") == "true"

	stats := []CacheStats{}

	// BigWig Data Cache
	wigDataStats := calculateRangeDataCacheSize(
		"bigwig-data",
		bigwig.BigWigDataCache,
		includeKeys,
		func(data bigwig.BigWigData) int64 {
			// BigWigData: Chr(string) + Start(int32) + End(int32) + Value(float32)
			return int64(len(data.Chr)) + 4 + 4 + 4
		},
	)
	stats = append(stats, wigDataStats)

	// BigWig Header Cache
	wigHeaderStats := calculateHeaderCacheSize(
		"bigwig-headers",
		bigwig.BigWigHeaderCache,
		includeKeys,
	)
	stats = append(stats, wigHeaderStats)

	// BigBed Data Cache
	bedDataStats := calculateRangeDataCacheSize(
		"bigbed-data",
		bigbed.BigBedDataCache,
		includeKeys,
		func(data bigbed.BigBedData) int64 {
			// BigBedData: Chr(string) + Start(int32) + End(int32) + Rest(string)
			return int64(len(data.Chr)) + 4 + 4 + int64(len(data.Rest))
		},
	)
	stats = append(stats, bedDataStats)

	// BigBed Header Cache
	bedHeaderStats := calculateHeaderCacheSize(
		"bigbed-headers",
		bigbed.BigBedHeaderCache,
		includeKeys,
	)
	stats = append(stats, bedHeaderStats)

	// Calculate totals
	var totalKB int64
	for _, stat := range stats {
		totalKB += stat.ApproxSizeKB
	}

	// Get runtime memory stats
	includeRuntime := r.URL.Query().Get("runtime") == "true"
	var memStats *MemoryStats
	if includeRuntime {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		memStats = &MemoryStats{
			AllocMB:      m.Alloc / 1024 / 1024,
			TotalAllocMB: m.TotalAlloc / 1024 / 1024,
			SysMB:        m.Sys / 1024 / 1024,
			NumGC:        m.NumGC,
			HeapAllocMB:  m.HeapAlloc / 1024 / 1024,
			HeapSysMB:    m.HeapSys / 1024 / 1024,
			HeapIdleMB:   m.HeapIdle / 1024 / 1024,
			HeapInUseMB:  m.HeapInuse / 1024 / 1024,
			StackInUseMB: m.StackInuse / 1024 / 1024,
		}
	}

	response := CacheStatusResponse{
		Caches:        stats,
		TotalSizeKB:   totalKB,
		TotalSizeMB:   totalKB / 1024,
		RuntimeMemory: memStats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// calculateHeaderCacheSize calculates the exact memory size of a BigData header cache
func calculateHeaderCacheSize(
	name string,
	c *cache.Cache[*bigdata.BigData],
	includeKeys bool,
) CacheStats {
	stats := CacheStats{
		Name:       name,
		EntryCount: c.Len(),
	}

	if includeKeys {
		stats.Keys = c.Keys()
	}

	var totalBytes int64

	// Iterate through all cached headers
	keys := c.Keys()
	for _, key := range keys {
		// Key string overhead
		totalBytes += int64(len(key))
		totalBytes += int64(unsafe.Sizeof(key)) // string header (16 bytes)

		// Get the cached BigData header
		bd, ok := c.Get(key)
		if !ok {
			continue
		}

		// Calculate size of BigData struct and its fields
		totalBytes += int64(unsafe.Sizeof(*bd)) // Base struct size

		// URL string
		totalBytes += int64(len(bd.URL))

		// AutoSql string
		totalBytes += int64(len(bd.AutoSql))

		// ZoomLevels slice
		totalBytes += int64(len(bd.ZoomLevels)) * int64(unsafe.Sizeof(bigdata.ZoomLevelHeader{}))

		// ChromTree maps
		for chromName, chromID := range bd.ChromTree.ChromToID {
			totalBytes += int64(len(chromName))     // key string
			totalBytes += int64(unsafe.Sizeof(key)) // string header
			totalBytes += int64(unsafe.Sizeof(chromID))
		}

		for chromName, chromSize := range bd.ChromTree.ChromSize {
			totalBytes += int64(len(chromName))
			totalBytes += int64(unsafe.Sizeof(key))
			totalBytes += int64(unsafe.Sizeof(chromSize))
		}

		for chromID, chromName := range bd.ChromTree.IDToChrom {
			totalBytes += int64(unsafe.Sizeof(chromID))
			totalBytes += int64(len(chromName))
			totalBytes += int64(unsafe.Sizeof(key))
		}
	}

	stats.ApproxSizeKB = totalBytes / 1024
	stats.ApproxSizeMB = stats.ApproxSizeKB / 1024

	return stats
}

// calculateRangeDataCacheSize calculates the approximate memory size of a RangeDataCache
func calculateRangeDataCacheSize[T any](
	name string,
	c *cache.RangeDataCache[T],
	includeKeys bool,
	itemSize func(T) int64,
) CacheStats {
	stats := CacheStats{
		Name:       name,
		EntryCount: c.Len(),
	}

	if includeKeys {
		stats.Keys = c.Keys()
	}

	var totalBytes int64

	// Iterate through all cache entries
	keys := c.Keys()
	for _, key := range keys {
		// Key string overhead
		totalBytes += int64(len(key))
		totalBytes += int64(unsafe.Sizeof(key)) // string header

		// Get the cached range data
		rangeDataList, ok := c.Get(key)
		if !ok {
			continue
		}

		// Calculate size of []RangeData[T]
		for _, rangeData := range rangeDataList {
			// RangeData struct overhead: Start(int) + End(int) + Data([]T)
			totalBytes += int64(unsafe.Sizeof(rangeData.Start))
			totalBytes += int64(unsafe.Sizeof(rangeData.End))
			totalBytes += int64(unsafe.Sizeof(rangeData.Data)) // slice header

			// Calculate size of actual data items
			for _, item := range rangeData.Data {
				totalBytes += itemSize(item)
			}
		}

		// Slice overhead for []RangeData[T]
		totalBytes += int64(unsafe.Sizeof(rangeDataList))
	}

	stats.ApproxSizeKB = totalBytes / 1024
	stats.ApproxSizeMB = stats.ApproxSizeKB / 1024

	return stats
}
