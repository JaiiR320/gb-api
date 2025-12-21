package bigwig

import (
	"errors"
	"gb-api/track/bigdata"
)

const (
	BIGWIG_MAGIC_LTH = 0x888FFC26 // BigWig Magic Low to High
	BIGWIG_MAGIC_HTL = 0x26FC8F88 // BigWig Magic High to Low
)

type BigWig struct {
	*bigdata.BigData
}

// BigWigData represents a single data point in a BigWig file
type BigWigData struct {
	Chr   string  `json:"chr"`
	Start int32   `json:"start"`
	End   int32   `json:"end"`
	Value float32 `json:"value"`
}

func ReadBigWig(url string, chr string, start int, end int) ([]BigWigData, error) {
	bw := BigWig{
		BigData: &bigdata.BigData{URL: url},
	}

	err := bw.LoadHeader(BIGWIG_MAGIC_LTH, BIGWIG_MAGIC_HTL)
	if err != nil {
		return nil, errors.New("Failed to load BigBed header: " + err.Error())
	}

	err = bw.LoadMetaData()
	if err != nil {
		return nil, errors.New("Failed to load metadata: " + err.Error())
	}

	data, err := bigdata.ReadData(bw.BigData, chr, int32(start), int32(end), decodeWigData)
	if err != nil {
		return nil, errors.New("Failed to read BigWig data: " + err.Error())
	}
	return data, nil
}
