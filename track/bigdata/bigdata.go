package bigdata

import (
	"encoding/binary"
	"errors"
)

const (
	BBFILE_HEADER_SIZE = 64 // Common header size
)

type BigData struct {
	URL          string            `json:"url"`
	Header       Header            `json:"header"`
	ZoomLevels   []ZoomLevelHeader `json:"zoomLevels"`
	ByteOrder    binary.ByteOrder  `json:"-"`
	AutoSql      string            `json:"autoSql,omitempty"`
	TotalSummary TotalSummary      `json:"totalSummary"`
	ChromTree    ChromTree         `json:"chromTree"`
	LTH          uint32            `json:"lowToHigh"`
	HTL          uint32            `json:"highToLow"`
}

type Header struct {
	Version            uint16 `json:"version"`
	NZoomLevels        uint16 `json:"nZoomLevels"`
	ChromTreeOffset    uint64 `json:"chromTreeOffset"`
	FullDataOffset     uint64 `json:"fullDataOffset"`
	FullIndexOffset    uint64 `json:"fullIndexOffset"`
	FieldCount         uint16 `json:"fieldCount"`
	DefinedFieldCount  uint16 `json:"definedFieldCount"`
	AutoSqlOffset      uint64 `json:"autoSqlOffset"`
	TotalSummaryOffset uint64 `json:"totalSummaryOffset"`
	UncompressBuffSize int32  `json:"uncompressBuffSize"`
	Reserved           uint64 `json:"reserved"`
}

type ZoomLevelHeader struct {
	Index          int    `json:"index"`
	ReductionLevel int32  `json:"reductionLevel"`
	Reserved       int32  `json:"reserved"`
	DataOffset     uint64 `json:"dataOffset"`
	IndexOffset    uint64 `json:"indexOffset"`
}

type TotalSummary struct {
	BasesCovered uint64  `json:"basesCovered"`
	MinVal       float64 `json:"minVal"`
	MaxVal       float64 `json:"maxVal"`
	SumData      float64 `json:"sumData"`
	SumSquares   float64 `json:"sumSquares"`
}

func New(url string, lth uint32, htl uint32) (*BigData, error) {
	b := BigData{URL: url, LTH: lth, HTL: htl}
	err := b.LoadHeader()
	if err != nil {
		return nil, err
	}

	err = b.LoadMetaData()
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func ReadBigData[T any](b *BigData, chr string, start int, end int, decode DataDecoder[T]) ([]T, error) {
	data, err := ReadData(b, chr, int32(start), int32(end), decode)
	if err != nil {
		return nil, errors.New("Failed to read BigWig data: " + err.Error())
	}
	return data, nil
}

// SelectZoomLevel returns the optimal zoom level index based on bases per pixel.
// Returns -1 if full resolution should be used, or the index into ZoomLevels array.
//
// Algorithm:
// - basesPerPixel = (end - start) / preRenderedWidth
// - Select the zoom level with reductionLevel closest to but NOT exceeding basesPerPixel
// - Use full resolution if basesPerPixel < threshold or no zoom levels exist
func (b *BigData) SelectZoomLevel(start, end int, preRenderedWidth int) int {
	// Guard against invalid inputs
	if preRenderedWidth <= 0 || len(b.ZoomLevels) == 0 {
		return -1 // Use full resolution
	}

	basesPerPixel := float64(end-start) / float64(preRenderedWidth)

	// Threshold: only use zoom if we can save at least 2x (configurable)
	const ZOOM_THRESHOLD = 2.0
	if basesPerPixel < ZOOM_THRESHOLD {
		return -1 // Full resolution is fine
	}

	// Find best zoom level: highest reductionLevel that doesn't exceed basesPerPixel
	bestZoomIdx := -1
	bestReductionLevel := int32(0)

	for i, zoom := range b.ZoomLevels {
		// Skip zoom levels with reduction > basesPerPixel (too coarse)
		if float64(zoom.ReductionLevel) > basesPerPixel {
			continue
		}

		// Pick the highest reduction level that fits
		if zoom.ReductionLevel > bestReductionLevel {
			bestReductionLevel = zoom.ReductionLevel
			bestZoomIdx = i
		}
	}

	return bestZoomIdx
}
