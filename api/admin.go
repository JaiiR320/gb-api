package api

import (
	"encoding/json"
	"gb-api/cache"
	"gb-api/track/bigdata/bigbed"
	"gb-api/track/bigdata/bigwig"
	"net/http"
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
	Caches      []CacheStats `json:"caches"`
	TotalSizeKB int64        `json:"totalSizeKB"`
	TotalSizeMB int64        `json:"totalSizeMB"`
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
	wigHeaderStats := CacheStats{
		Name:       "bigwig-headers",
		EntryCount: bigwig.BigWigHeaderCache.Len(),
		// Headers contain bigdata.BigData structs - approximate size
		ApproxSizeKB: int64(bigwig.BigWigHeaderCache.Len() * 10), // ~10KB per header (rough estimate)
	}
	wigHeaderStats.ApproxSizeMB = wigHeaderStats.ApproxSizeKB / 1024
	if includeKeys {
		wigHeaderStats.Keys = bigwig.BigWigHeaderCache.Keys()
	}
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
	bedHeaderStats := CacheStats{
		Name:         "bigbed-headers",
		EntryCount:   bigbed.BigBedHeaderCache.Len(),
		ApproxSizeKB: int64(bigbed.BigBedHeaderCache.Len() * 10), // ~10KB per header
	}
	bedHeaderStats.ApproxSizeMB = bedHeaderStats.ApproxSizeKB / 1024
	if includeKeys {
		bedHeaderStats.Keys = bigbed.BigBedHeaderCache.Keys()
	}
	stats = append(stats, bedHeaderStats)

	// Calculate totals
	var totalKB int64
	for _, stat := range stats {
		totalKB += stat.ApproxSizeKB
	}

	response := CacheStatusResponse{
		Caches:      stats,
		TotalSizeKB: totalKB,
		TotalSizeMB: totalKB / 1024,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
