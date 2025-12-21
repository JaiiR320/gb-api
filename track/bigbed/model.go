package bigbed

import (
	"encoding/binary"
	"encoding/json"
)

type BigBed struct {
	URL          string            `json:"url"`
	Header       Header            `json:"header"`
	ZoomLevels   []ZoomLevelHeader `json:"zoomLevels"`
	ByteOrder    binary.ByteOrder  `json:"-"`
	AutoSql      string            `json:"autoSql,omitempty"`
	TotalSummary BBTotalSummary    `json:"totalSummary"`
	ChromTree    ChromTree         `json:"chromTree"`
}

type Header struct {
	BbVersion          uint16 `json:"bbVersion"`
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

type BBTotalSummary struct {
	BasesCovered uint64  `json:"basesCovered"`
	MinVal       float64 `json:"minVal"`
	MaxVal       float64 `json:"maxVal"`
	SumData      float64 `json:"sumData"`
	SumSquares   float64 `json:"sumSquares"`
}

type ChromTree struct {
	BlockSize int32            `json:"blockSize"`
	KeySize   int32            `json:"keySize"`
	ValSize   int32            `json:"valSize"`
	ItemCount uint64           `json:"itemCount"`
	Reserved  uint64           `json:"reserved"`
	ChromToID map[string]int32 `json:"chromToId"`
	ChromSize map[string]int32 `json:"chromSize"`
	IDToChrom map[int32]string `json:"idToChrom"`
}

// MarshalJSON implements custom JSON marshaling for BigBed
func (b *BigBed) MarshalJSON() ([]byte, error) {
	type Alias BigBed
	var byteOrderStr string
	byteOrderStr = "LittleEndian"
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

// BigBedData represents a single data point in a BigBed file
type BigBedData struct {
	Chr   string `json:"chr"`
	Start int32  `json:"start"`
	End   int32  `json:"end"`
	Name  string `json:"name,omitempty"`
	Score int32  `json:"score,omitempty"`
	Rest  string `json:"rest,omitempty"`
}
