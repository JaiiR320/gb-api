package bigbed

import (
	"errors"
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

func ReadBigBed(url string, chr string, start int, end int) ([]BigBedData, error) {
	bb := BigBed{
		BigData: &bigdata.BigData{URL: url},
	}

	err := bb.LoadHeader(BIGBED_MAGIC_LTH, BIGBED_MAGIC_HTL)
	if err != nil {
		return nil, errors.New("Failed to load BigBed header: " + err.Error())
	}

	err = bb.LoadMetaData()
	if err != nil {
		return nil, errors.New("Failed to load metadata: " + err.Error())
	}

	data, err := bigdata.ReadData(bb.BigData, chr, int32(start), int32(end), decodeBedData)
	if err != nil {
		return nil, errors.New("Failed to read BigWig data: " + err.Error())
	}
	return data, nil
}
