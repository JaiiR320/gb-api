package bigbed

import (
	"bytes"
	"gb-api/track/bigdata"
	"gb-api/utils"
)

func decodeBedData(b *bigdata.BigData, data []byte, filterStartChromIndex int32, filterStartBase int32,
	filterEndChromIndex int32, filterEndBase int32) ([]BigBedData, error) {

	decodedData := []BigBedData{}
	binaryParser := utils.NewParser(bytes.NewReader(data), b.ByteOrder)

	for {
		// Read chromIndex
		chromIndex, err := binaryParser.GetInt32()
		if err != nil {
			break
		}

		// Read startBase
		startBase, err := binaryParser.GetInt32()
		if err != nil {
			break
		}

		// Read endBase
		endBase, err := binaryParser.GetInt32()
		if err != nil {
			break
		}

		// Read rest string (null-terminated)
		rest, err := binaryParser.GetString(0) // 0 means read until null terminator
		if err != nil {
			break
		}

		// Get chromosome name
		if chromIndex < 0 || int(chromIndex) >= len(b.ChromTree.IDToChrom) {
			continue
		}
		chrom := b.ChromTree.IDToChrom[chromIndex]

		// Filter: skip entries before the start
		if chromIndex < filterStartChromIndex || (chromIndex == filterStartChromIndex && endBase < filterStartBase) {
			continue
		}

		// Filter: stop processing entries after the end
		if chromIndex > filterEndChromIndex || (chromIndex == filterEndChromIndex && startBase >= filterEndBase) {
			break
		}

		// Create entry
		entry := BigBedData{
			Chr:   chrom,
			Start: startBase,
			End:   endBase,
			Rest:  rest,
		}
		decodedData = append(decodedData, entry)
	}

	return decodedData, nil
}
