package bigbed

type CCRE struct {
	Chr    string `json:"chr"`
	Start  int32  `json:"start"`
	End    int32  `json:"end"`
	Name   string `json:"name"`
	Score  int32  `json:"score"`
	Strand string `json:"strans"`

	Rest string `json:"rest,omitempty"`
}

func ParseCCRE(data []BigBedData) (any, error) {
	return nil, nil
}
