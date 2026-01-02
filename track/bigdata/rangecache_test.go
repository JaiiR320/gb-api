package bigdata

import (
	"reflect"
	"testing"
)

type Test struct {
	name     string
	start    int
	end      int
	cached   []RangeData[struct{}]
	expected []Range
}

var tests = []Test{
	{
		name:     "left of cached ranges",
		start:    0,
		end:      100,
		cached:   []RangeData[struct{}]{{Start: 200, End: 300}},
		expected: []Range{{Start: 0, End: 100}},
	},
	{
		name:     "start outside, end inside first cache",
		start:    0,
		end:      100,
		cached:   []RangeData[struct{}]{{Start: 50, End: 150}},
		expected: []Range{{Start: 0, End: 50}},
	},
	{
		name:     "start and end inside first cache",
		start:    75,
		end:      125,
		cached:   []RangeData[struct{}]{{Start: 50, End: 150}},
		expected: []Range{},
	},
	{
		name:     "start inside, end outside first cache",
		start:    75,
		end:      200,
		cached:   []RangeData[struct{}]{{Start: 50, End: 150}},
		expected: []Range{{Start: 150, End: 200}},
	},
	{
		name:     "start and end right of first cache",
		start:    175,
		end:      200,
		cached:   []RangeData[struct{}]{{Start: 50, End: 150}},
		expected: []Range{{175, 200}},
	},
	{
		name:     "start and end are same as cached",
		start:    100,
		end:      200,
		cached:   []RangeData[struct{}]{{Start: 100, End: 200}},
		expected: []Range{},
	},
	{
		name:     "start and end are same as cached",
		start:    100,
		end:      200,
		cached:   []RangeData[struct{}]{{Start: 100, End: 200}},
		expected: []Range{},
	},
}

func TestRangeCache(t *testing.T) {
	for _, test := range tests {
		res := FindRanges(test.start, test.end, test.cached)
		if !reflect.DeepEqual(res, test.expected) {
			t.Errorf("test %s: got %v, wanted %v", test.name, res, test.expected)
		}
	}
}
