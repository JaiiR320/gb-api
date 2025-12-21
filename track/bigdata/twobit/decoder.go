package twobit

import (
	"bytes"
	"gb-api/track/bigdata"
	"gb-api/utils"
)

func decodeWigData(b *bigdata.BigData, data []byte, filterStartChromIndex int32, filterStartBase int32,
	filterEndChromIndex int32, filterEndBase int32) ([]TwoBitData, error) {
	_ = []TwoBitData{}
	_ = utils.NewParser(bytes.NewReader(data), b.ByteOrder)
	return nil, nil
}
