package bigwig

import (
	"fmt"
	"gb-api/cache"
	"gb-api/track/bigdata"
	"sort"
	"sync"
)

// create a cache for headers, and for data ranges
var BigWigHeaderCache *cache.Cache[*bigdata.BigData]
var BigWigDataCache *cache.RangeDataCache[BigWigData]

func init() {
	dataCache, err := cache.NewCache[[]cache.RangeData[BigWigData]](25)
	if err != nil {
		panic(err)
	}
	BigWigDataCache = dataCache

	headerCache, err := cache.NewCache[*bigdata.BigData](25)
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

	fmt.Printf("[Cache] Request: url=%s, chrom=%s, start=%d, end=%d, zoomIdx=%d\n",
		url, chrom, start, end, zoomIdx)

	// ranges start out as original request
	rangesToFetch := []cache.Range{{Start: start, End: end}}

	// generate new ranges on cache hit
	cachedData, hit := BigWigDataCache.Get(cacheId)
	if hit {
		fmt.Printf("[Cache] HIT! Found %d cached ranges for zoom level %d\n", len(cachedData), zoomIdx)
		rangesToFetch = cache.FindRanges(start, end, cachedData)
		fmt.Printf("[Cache] Need to fetch %d ranges: %v\n", len(rangesToFetch), rangesToFetch)
	} else {
		fmt.Printf("[Cache] MISS! Need to fetch entire range\n")
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

	var erra error

	for _, r := range rangesToFetch {
		go func(r cache.Range) {
			defer wg.Done()
			fmt.Printf("[Cache] Goroutine fetching: start=%d, end=%d, zoom=%d\n", r.Start, r.End, zoomIdx)

			data, err := bigdata.ReadDataWithZoom(bw, chrom, int32(r.Start), int32(r.End),
				decoder, zoomIdx)
			if err != nil {
				erra = err
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

	// Merge cached and newly fetched data
	rangeData := cachedData
	for dc := range dchan {
		rangeData = append(rangeData, dc)
	}

	fmt.Printf("[Cache] Collected %d total ranges\n", len(rangeData))

	sort.Slice(rangeData, func(i, j int) bool {
		return rangeData[i].Start < rangeData[j].Start
	})

	// Merge overlapping/adjacent ranges
	rangeData = cache.MergeRanges(rangeData)
	fmt.Printf("[Cache] After merging: %d ranges\n", len(rangeData))

	BigWigDataCache.Add(cacheId, rangeData)

	// Filter data to only include points within the requested range
	var data []BigWigData
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
	fmt.Printf("[Cache] Returning %d data points\n", len(data))

	return data, erra
}
