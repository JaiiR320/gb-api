package bigbed

import (
	"fmt"
	"gb-api/track/bigdata"
)

const (
	BIGBED_MAGIC_LTH = 0x8789F2EB // BigBed Magic Low to High
	BIGBED_MAGIC_HTL = 0xEBF28987 // BigBed Magic High to Low
)

type BigBed struct {
	*bigdata.BigData
}

// BigBedData represents a single data point in a BigBed file
type BigBedData struct {
	Chr   string `json:"chr"`
	Start int32  `json:"start"`
	End   int32  `json:"end"`
	Rest  string `json:"rest,omitempty"`
}

// ReadBigBed reads data without caching (use GetCachedBedData for cached reads)
func ReadBigBed(url string, chr string, start int, end int) ([]BigBedData, error) {
	bb, err := bigdata.New(url, BIGBED_MAGIC_LTH, BIGBED_MAGIC_HTL)
	if err != nil {
		return nil, fmt.Errorf("Failed to create bigbed, %w", err)
	}
	data, err := bigdata.ReadData(bb, chr, int32(start), int32(end), decodeBedData)
	if err != nil {
		return nil, fmt.Errorf("Failed to read BigBed data, %w", err)
	}
	return data, nil
}
