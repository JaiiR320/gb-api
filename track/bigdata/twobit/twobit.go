package twobit

import "gb-api/track/bigdata"

const (
	TWOBIT_MAGIC_LTH = 0x1A412743 // BigWig Magic High to Low
	TWOBIT_MAGIC_HTL = 0x4327411A // BigWig Magic Low to High
)

type TwoBit struct {
	*bigdata.BigData
}

// BigWigData represents a single data point in a BigWig file
type TwoBitData struct {
	Chr   string `json:"chr"`
	Start int32  `json:"start"`
	End   int32  `json:"end"`
	//...
}
