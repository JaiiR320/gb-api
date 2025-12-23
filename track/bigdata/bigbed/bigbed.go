package bigbed

import (
	"fmt"
	"gb-api/track/bigdata"
)

const (
	BIGBED_MAGIC_LTH = 0x8789F2EB // BigBed Magic Low to High
	BIGBED_MAGIC_HTL = 0xEBF28987 // BigBed Magic High to Low       = 0x2468ACE0
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

var BigBedCache *bigdata.Cache

func init() {
	BigBedCache = bigdata.NewCache()
}

func getBigBed(url string) (*bigdata.BigData, error) {
	if cached, ok := BigBedCache.Get(url); ok {
		return cached, nil
	}
	bw, err := bigdata.New(url, BIGBED_MAGIC_LTH, BIGBED_MAGIC_HTL)
	if err != nil {
		return nil, err
	}

	BigBedCache.Set(url, bw)
	return bw, nil
}

func ReadBigBed(url string, chr string, start int, end int) ([]BigBedData, error) {
	bb, err := getBigBed(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to create bigbed, %w", err)
	}
	data, err := bigdata.ReadData(bb, chr, int32(start), int32(end), decodeBedData)
	if err != nil {
		return nil, fmt.Errorf("Failed to read BigWig data, %w", err)
	}
	return data, nil
}
