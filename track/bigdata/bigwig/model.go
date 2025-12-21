package bigwig

import "gb-api/track/bigdata"

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
