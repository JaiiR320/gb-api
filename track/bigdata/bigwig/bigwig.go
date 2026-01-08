package bigwig

import (
	"fmt"
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

// simple implementation, no caching
func ReadBigWig(url string, chr string, start int, end int, preRenderedWidth int) ([]BigWigData, error) {
	bw, err := bigdata.New(url, BIGWIG_MAGIC_LTH, BIGWIG_MAGIC_HTL)
	if err != nil {
		return nil, err
	}

	// Select optimal zoom level
	zoomIdx := bw.SelectZoomLevel(start, end, preRenderedWidth)

	// Select appropriate decoder
	var decoder bigdata.DataDecoder[BigWigData]
	if zoomIdx >= 0 {
		decoder = decodeZoomData
	} else {
		decoder = decodeWigData
	}

	data, err := bigdata.ReadDataWithZoom(bw, chr, int32(start), int32(end), decoder, zoomIdx)
	if err != nil {
		return nil, fmt.Errorf("Failed to read BigWig data, %w", err)
	}

	return data, nil
}
