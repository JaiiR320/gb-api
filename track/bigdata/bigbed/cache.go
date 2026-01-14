package bigbed

import (
	"fmt"
	"gb-api/cache"
	"gb-api/config"
	"gb-api/track/bigdata"
	"log/slog"
	"sort"
	"sync"
)

// create a cache for headers, and for data ranges
var BigBedHeaderCache *cache.Cache[*bigdata.BigData]
var BigBedDataCache *cache.RangeDataCache[BigBedData]

func init() {
	cacheSize := config.GetCacheSize()

	dataCache, err := cache.NewCache[[]cache.RangeData[BigBedData]](cacheSize)
	if err != nil {
		panic(err)
	}
	BigBedDataCache = dataCache

	headerCache, err := cache.NewCache[*bigdata.BigData](cacheSize)
	if err != nil {
		panic(err)
	}
	BigBedHeaderCache = headerCache
}

func getCachedHeader(url string) (*bigdata.BigData, error) {
	if cached, ok := BigBedHeaderCache.Get(url); ok {
		return cached, nil
	}
	bb, err := bigdata.New(url, BIGBED_MAGIC_LTH, BIGBED_MAGIC_HTL)
	if err != nil {
		return nil, err
	}

	BigBedHeaderCache.Add(url, bb)
	return bb, nil
}

func GetCachedBedData(url string, chrom string, start, end int) ([]BigBedData, error) {
	slog.Debug("Cache request", "url", url, "chrom", chrom, "start", start, "end", end)
	cacheId := url + "-" + chrom
	// ranges start out as original request
	rangesToFetch := []cache.Range{{Start: start, End: end}}

	// generate new ranges on cache hit
	cachedData, hit := BigBedDataCache.Get(cacheId)
	if hit {
		slog.Debug("Cache hit", "cachedRanges", len(cachedData))
		rangesToFetch = cache.FindRanges(start, end, cachedData)
		slog.Debug("Ranges to fetch", "count", len(rangesToFetch), "ranges", rangesToFetch)
	} else {
		slog.Debug("Cache miss", "fetchingEntireRange", true)
	}

	// stupid race condition temporary solution
	var erra error

	dchan := make(chan cache.RangeData[BigBedData], len(rangesToFetch))

	var wg sync.WaitGroup
	wg.Add(len(rangesToFetch))

	bb, err := getCachedHeader(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to create bigbed, %w", err)
	}

	for _, r := range rangesToFetch {
		go func(r cache.Range) {
			defer wg.Done()
			slog.Debug("Goroutine fetching", "start", r.Start, "end", r.End)
			data, err := bigdata.ReadData(bb, chrom, int32(r.Start), int32(r.End), decodeBedData)
			if err != nil {
				erra = err
			}

			rdata := cache.RangeData[BigBedData]{
				Start: r.Start,
				End:   r.End,
				Data:  data,
			}
			dchan <- rdata
		}(r)
	}
	wg.Wait()
	close(dchan)

	rangeData := cachedData
	for dc := range dchan {
		rangeData = append(rangeData, dc)
	}
	slog.Debug("Collected ranges", "total", len(rangeData))

	sort.Slice(rangeData, func(i, j int) bool {
		return rangeData[i].Start < rangeData[j].Start
	})

	// Merge overlapping/adjacent ranges
	rangeData = cache.MergeRanges(rangeData)
	slog.Debug("After merging", "ranges", len(rangeData))

	BigBedDataCache.Add(cacheId, rangeData)

	// Filter data to only include points within the requested range
	var data []BigBedData
	for _, r := range rangeData {
		// Only include data from ranges that overlap with the requested region
		if r.End <= start || r.Start >= end {
			continue // Range doesn't overlap with request
		}

		// Add data points that fall within the requested range
		for _, point := range r.Data {
			if point.Start >= int32(start) && point.Start < int32(end) {
				data = append(data, point)
			}
		}
	}
	slog.Debug("Returning data points", "count", len(data), "start", start, "end", end)

	return data, erra
}
