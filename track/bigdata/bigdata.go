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
