package bigbed

import "gb-api/track/bigdata"

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
