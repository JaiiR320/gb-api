package bigdata

import (
	"encoding/binary"
)

type BigData struct {
	URL          string            `json:"url"`
	Header       Header            `json:"header"`
	ZoomLevels   []ZoomLevelHeader `json:"zoomLevels"`
	ByteOrder    binary.ByteOrder  `json:"-"`
	AutoSql      string            `json:"autoSql,omitempty"`
	TotalSummary TotalSummary      `json:"totalSummary"`
	ChromTree    ChromTree         `json:"chromTree"`
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
