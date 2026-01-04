package bigdata

type Range struct {
	Start int
	End   int
}

type RangeData[T any] struct {
	Start int
	End   int
	Data  []T
}

type RangeDataCache[T any] = Cache[[]RangeData[T]]

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

func MergeRanges[T any](ranges []RangeData[T]) []RangeData[T] {
	if len(ranges) == 0 {
		return ranges
	}

	result := []RangeData[T]{}
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
