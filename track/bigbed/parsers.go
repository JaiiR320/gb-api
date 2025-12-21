package bigbed

import (
	"errors"
	"strconv"
	"strings"
)

type CCRE struct {
	BigBedData
	Name       string `json:"name"`
	Score      int32  `json:"score"`
	Strand     string `json:"strand"`
	ThickStart int32  `json:"thickStart"`
	ThickEnd   int32  `json:"thickEnd"`
	Color      string `json:"color"`
	Class      string `json:"class"`
}

const lenFields = 7

func ParseCCRE(data []BigBedData) ([]CCRE, error) {
	var out = make([]CCRE, len(data))
	for i, d := range data {
		fields := strings.Split(d.Rest, "\t")

		if len(fields) < lenFields {
			return nil, errors.New("Incorrect number of fields in rest")
		}

		name := fields[0]
		score, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, errors.New("Unable to parse thickstart")
		}
		strand := fields[2]
		thickstart, err := strconv.Atoi(fields[3])
		if err != nil {
			return nil, errors.New("Unable to parse thickstart")
		}
		thickend, err := strconv.Atoi(fields[4])
		if err != nil {
			return nil, errors.New("Unable to parse thickend")
		}
		color := fields[5]
		class := fields[6]

		d.Rest = strings.Join(fields[lenFields:], "\t")

		out[i] = CCRE{
			BigBedData: d,
			Name:       name,
			Score:      int32(score),
			Strand:     strand,
			ThickStart: int32(thickstart),
			ThickEnd:   int32(thickend),
			Color:      color,
			Class:      class,
		}
	}
	return out, nil
}
