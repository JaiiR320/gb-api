package bigwig

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
var BigWigHeaderCache *cache.Cache[*bigdata.BigData]
var BigWigDataCache *cache.RangeDataCache[BigWigData]

func init() {
	cacheSize := config.GetCacheSize()

	dataCache, err := cache.NewCache[[]cache.RangeData[BigWigData]](cacheSize)
	if err != nil {
		panic(err)
	}
	BigWigDataCache = dataCache

	headerCache, err := cache.NewCache[*bigdata.BigData](cacheSize)
	if err != nil {
		panic(err)
	}
	BigWigHeaderCache = headerCache
}

func getCachedHeader(url string) (*bigdata.BigData, error) {
	if cached, ok := BigWigHeaderCache.Get(url); ok {
		return cached, nil
	}
	bw, err := bigdata.New(url, BIGWIG_MAGIC_LTH, BIGWIG_MAGIC_HTL)
	if err != nil {
		return nil, err
	}

	BigWigHeaderCache.Add(url, bw)
	return bw, nil
}

func GetCachedWigData(url string, chrom string, start, end int, preRenderedWidth int) ([]BigWigData, error) {
	bw, err := getCachedHeader(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to create bigwig, %w", err)
	}

	// Select optimal zoom level
	zoomIdx := bw.SelectZoomLevel(start, end, preRenderedWidth)

	// Create cache key that includes zoom level
	var cacheId string
	if zoomIdx >= 0 {
		cacheId = fmt.Sprintf("%s-%s-zoom%d", url, chrom, zoomIdx)
	} else {
		cacheId = fmt.Sprintf("%s-%s", url, chrom)
	}

	slog.Debug("Cache request", "url", url, "chrom", chrom, "start", start, "end", end, "zoomIdx", zoomIdx)

	// ranges start out as original request
	rangesToFetch := []cache.Range{{Start: start, End: end}}

	// generate new ranges on cache hit
	cachedData, hit := BigWigDataCache.Get(cacheId)
	if hit {
		slog.Debug("Cache hit", "cachedRanges", len(cachedData), "zoomIdx", zoomIdx)
		rangesToFetch = cache.FindRanges(start, end, cachedData)
		slog.Debug("Ranges to fetch", "count", len(rangesToFetch), "ranges", rangesToFetch)
	} else {
		slog.Debug("Cache miss", "fetchingEntireRange", true)
	}

	// Select appropriate decoder
	var decoder bigdata.DataDecoder[BigWigData]
	if zoomIdx >= 0 {
		decoder = decodeZoomData
	} else {
		decoder = decodeWigData
	}

	dchan := make(chan cache.RangeData[BigWigData], len(rangesToFetch))
	var wg sync.WaitGroup
	wg.Add(len(rangesToFetch))

	errchan := make(chan error, 1)

	for _, r := range rangesToFetch {
		go func(r cache.Range) {
			defer wg.Done()
			slog.Debug("Goroutine fetching", "start", r.Start, "end", r.End, "zoomIdx", zoomIdx)

			data, err := bigdata.ReadDataWithZoom(bw, chrom, int32(r.Start), int32(r.End),
				decoder, zoomIdx)
			if err != nil {
				select {
				case errchan <- err:
				default:
				}
				return
			}

			dchan <- cache.RangeData[BigWigData]{
				Start: r.Start,
				End:   r.End,
				Data:  data,
			}
		}(r)
	}

	wg.Wait()
	close(dchan)
	close(errchan)

	// Check for errors from goroutines
	if err, ok := <-errchan; ok {
		return nil, err
	}

	// Merge cached and newly fetched data
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

	BigWigDataCache.Add(cacheId, rangeData)

	// Filter data to only include points within the requested range
	// Count total points for pre-allocation
	totalPoints := 0
	for _, r := range rangeData {
		totalPoints += len(r.Data)
	}
	data := make([]BigWigData, 0, totalPoints)
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
	slog.Debug("Returning data points", "count", len(data))

	return data, nil
}
