package bigwig

import (
	"encoding/binary"
	"encoding/json"
	"gb-api/track/common"
)

type BigWig struct {
	URL          string            `json:"url"`
	Header       Header            `json:"header"`
	ZoomLevels   []ZoomLevelHeader `json:"zoomLevels"`
	ByteOrder    binary.ByteOrder  `json:"-"`
	AutoSql      string            `json:"autoSql,omitempty"`
	TotalSummary BWTotalSummary    `json:"totalSummary"`
	ChromTree    common.ChromTree  `json:"chromTree"`
}

type Header struct {
	BwVersion          uint16 `json:"bwVersion"`
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

type BWTotalSummary struct {
	BasesCovered uint64  `json:"basesCovered"`
	MinVal       float64 `json:"minVal"`
	MaxVal       float64 `json:"maxVal"`
	SumData      float64 `json:"sumData"`
	SumSquares   float64 `json:"sumSquares"`
}

// MarshalJSON implements custom JSON marshaling for BigWig
func (b *BigWig) MarshalJSON() ([]byte, error) {
	type Alias BigWig
	byteOrderStr := "LittleEndian"
	if b.ByteOrder == binary.BigEndian {
		byteOrderStr = "BigEndian"
	}

	return json.Marshal(&struct {
		*Alias
		ByteOrder string `json:"byteOrder"`
	}{
		Alias:     (*Alias)(b),
		ByteOrder: byteOrderStr,
	})
}

// ToJSON returns a pretty-printed JSON representation of the BigWig struct
func (b *BigWig) ToJSON() (string, error) {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// BigWigData represents a single data point in a BigWig file
type BigWigData struct {
	Chr   string  `json:"chr"`
	Start int32   `json:"start"`
	End   int32   `json:"end"`
	Value float32 `json:"value"`
}
