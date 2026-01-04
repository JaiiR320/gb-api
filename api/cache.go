package api

import (
	"fmt"
	"gb-api/track/bigdata"
	"gb-api/track/bigdata/bigwig"
	"sort"
	"sync"
)

type BigWigCache struct {
	*bigdata.RangeCache[bigwig.BigWigData]
}

var WigCache *BigWigCache

func init() {
	cache, err := bigdata.NewRangeCache[bigwig.BigWigData](25)
	if err != nil {
		panic(err)
	}
	WigCache = &BigWigCache{cache}
}

func (b BigWigCache) GetCachedWigData(url string, chrom string, start, end int) ([]bigwig.BigWigData, error) {
	fmt.Printf("[Cache] Request: url=%s, chrom=%s, start=%d, end=%d\n", url, chrom, start, end)
	cacheId := url + "-" + chrom
	// ranges start out as original request
	rangesToFetch := []bigdata.Range{{Start: start, End: end}}

	// generate new ranges on cache hit
	cachedData, hit := b.Get(cacheId)
	if hit {
		fmt.Printf("[Cache] HIT! Found %d cached ranges\n", len(cachedData))
		rangesToFetch = bigdata.FindRanges(start, end, cachedData)
		fmt.Printf("[Cache] Need to fetch %d ranges: %v\n", len(rangesToFetch), rangesToFetch)
	} else {
		fmt.Printf("[Cache] MISS! Need to fetch entire range\n")
	}

	// stupid race condition temporary solution
	var erra error

	dchan := make(chan bigdata.RangeData[bigwig.BigWigData], len(rangesToFetch))

	var wg sync.WaitGroup
	wg.Add(len(rangesToFetch))

	for _, r := range rangesToFetch {
		go func(r bigdata.Range) {
			defer wg.Done()
			fmt.Printf("[Cache] Goroutine fetching: start=%d, end=%d\n", r.Start, r.End)
			data, err := bigwig.ReadBigWig(url, chrom, r.Start, r.End)
			if err != nil {
				erra = err
			}

			rdata := bigdata.RangeData[bigwig.BigWigData]{
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
	fmt.Printf("[Cache] Collected %d total ranges (cached + fetched)\n", len(rangeData))

	sort.Slice(rangeData, func(i, j int) bool {
		return rangeData[i].Start < rangeData[j].Start
	})

	// Merge overlapping/adjacent ranges
	rangeData = mergeRanges[bigwig.BigWigData](rangeData)
	fmt.Printf("[Cache] After merging: %d ranges\n", len(rangeData))

	b.Add(cacheId, rangeData)

	var data []bigwig.BigWigData
	for _, r := range rangeData {
		data = append(data, r.Data...)
	}
	fmt.Printf("[Cache] Returning %d data points\n", len(data))

	return data, erra
}

func mergeRanges[T any](ranges []bigdata.RangeData[T]) []bigdata.RangeData[T] {
	if len(ranges) == 0 {
		return ranges
	}

	result := []bigdata.RangeData[T]{}
	current := ranges[0] // Start with first range

	for i := 1; i < len(ranges); i++ {
		next := ranges[i]

		if next.Start <= current.End { // Does next overlap/touch current?
			// Merge: extend current's end, append data
			if next.End > current.End {
				current.End = next.End
			}
			current.Data = append(current.Data, next.Data...)
		} else {
			// Gap found: save current, start new one
			result = append(result, current)
			current = next
		}
	}

	result = append(result, current)
	return result
}
