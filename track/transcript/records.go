package transcript

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/brentp/bix"
	"github.com/brentp/irelate/parsers"
)

type Position struct {
	chrom string
	start int
	end   int
}

type Record struct {
	Chrom      string
	Source     string
	Feature    string
	Start      int
	End        int
	Score      string
	Strand     string
	Frame      string
	Attributes map[string]string
}

func (p Position) Chrom() string {
	return p.chrom
}

func (p Position) Start() uint32 {
	return uint32(p.start)
}

func (p Position) End() uint32 {
	return uint32(p.end)
}

func NewPosition(pos string) (Position, error) {
	left := strings.Split(pos, ":")
	if len(left) != 2 {
		return Position{}, fmt.Errorf("invalid position format: %s", pos)
	}
	right := strings.Split(left[1], "-")
	if len(right) != 2 {
		return Position{}, fmt.Errorf("invalid position range format: %s", left[1])
	}
	start, err := strconv.Atoi(strings.ReplaceAll(right[0], ",", ""))
	if err != nil {
		return Position{}, fmt.Errorf("invalid start position: %v", err)
	}
	end, err := strconv.Atoi(strings.ReplaceAll(right[1], ",", ""))
	if err != nil {
		return Position{}, fmt.Errorf("invalid end position: %v", err)
	}
	return Position{chrom: left[0], start: start, end: end}, nil
}

func GetRecords(pathStr string, posStr string) ([]Record, error) {
	tbx, err := bix.New(pathStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open tabix file: %v", err)
	}
	defer tbx.Close()

	pos, err := NewPosition(posStr)
	if err != nil {
		return nil, err
	}

	rdr, err := tbx.FastQuery(pos)
	if err != nil {
		return nil, fmt.Errorf("failed to query tabix file: %v", err)
	}
	defer rdr.Close()
	var records []Record
	for {
		line, err := rdr.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("error reading line: %v", err)
		}
		if line == nil {
			break
		}
		interval, ok := line.(*parsers.Interval)
		if !ok {
			return nil, fmt.Errorf("failed to cast line to Interval")
		}
		record, err := ParseRecord(interval)
		if err != nil {
			return nil, fmt.Errorf("error parsing record: %v", err)
		}

		records = append(records, record)
	}
	return records, nil
}

func ParseRecord(interval *parsers.Interval) (Record, error) {
	start, err := strconv.Atoi(string(interval.Fields[3]))
	if err != nil {
		return Record{}, fmt.Errorf("invalid start position: %v", err)
	}
	end, err := strconv.Atoi(string(interval.Fields[4]))
	if err != nil {
		return Record{}, fmt.Errorf("invalid end position: %v", err)
	}

	attributes := string(interval.Fields[8])
	var attrMap = make(map[string]string)
	attrPairs := strings.Split(attributes, ";")
	for _, pair := range attrPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, " ", 2)
		if len(kv) != 2 {
			return Record{}, fmt.Errorf("invalid attribute pair: %s", pair)
		}
		key := kv[0]
		value := strings.Trim(kv[1], `"`)
		if existing, ok := attrMap[key]; ok {
			attrMap[key] = existing + "," + value
		} else {
			attrMap[key] = value
		}
	}

	return Record{
		Chrom:      string(interval.Fields[0]),
		Source:     string(interval.Fields[1]),
		Feature:    string(interval.Fields[2]),
		Start:      start,
		End:        end,
		Score:      string(interval.Fields[5]),
		Strand:     string(interval.Fields[6]),
		Frame:      string(interval.Fields[7]),
		Attributes: attrMap,
	}, nil
}
