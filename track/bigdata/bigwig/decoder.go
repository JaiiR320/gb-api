package bigwig

import (
	"bytes"
	"gb-api/track/bigdata"
	"gb-api/utils"
)

func decodeWigData(b *bigdata.BigData, data []byte, filterStartChromIndex int32, filterStartBase int32,
	filterEndChromIndex int32, filterEndBase int32) ([]BigWigData, error) {

	decodedData := []BigWigData{}
	binaryParser := utils.NewParser(bytes.NewReader(data), b.ByteOrder)

	chromIndex, err := binaryParser.GetInt32()
	if err != nil {
		return nil, err
	}

	chrom := b.ChromTree.IDToChrom[chromIndex]

	startBase, err := binaryParser.GetInt32()
	if err != nil {
		return nil, err
	}

	endBase, err := binaryParser.GetInt32()
	if err != nil {
		return nil, err
	}

	itemStep, err := binaryParser.GetInt32()
	if err != nil {
		return nil, err
	}

	itemSpan, err := binaryParser.GetInt32()
	if err != nil {
		return nil, err
	}

	dataType, err := binaryParser.GetUInt8()
	if err != nil {
		return nil, err
	}

	// Skip reserved byte
	_, err = binaryParser.GetUInt8()
	if err != nil {
		return nil, err
	}

	itemCount, err := binaryParser.GetUInt16()
	if err != nil {
		return nil, err
	}

	if chromIndex < filterStartChromIndex || chromIndex > filterEndChromIndex {
		return decodedData, nil
	}

	for itemCount > 0 {
		itemCount--

		var value float32

		switch dataType {
		case 1:
			// Data is stored in Bed Graph format
			startBase, err = binaryParser.GetInt32()
			if err != nil {
				return nil, err
			}

			endBase, err = binaryParser.GetInt32()
			if err != nil {
				return nil, err
			}

			value, err = binaryParser.GetFloat32()
			if err != nil {
				return nil, err
			}
		case 2:
			// Data is stored in Variable Step format
			startBase, err = binaryParser.GetInt32()
			if err != nil {
				return nil, err
			}

			value, err = binaryParser.GetFloat32()
			if err != nil {
				return nil, err
			}

			endBase = startBase + itemSpan
		default:
			// Data is stored in Fixed Step format
			value, err = binaryParser.GetFloat32()
			if err != nil {
				return nil, err
			}

			endBase = startBase + itemSpan
		}

		// Check if past the end of the range; exit
		if chromIndex > filterEndChromIndex || (chromIndex == filterEndChromIndex && startBase >= filterEndBase) {
			break
		}

		// Check if within the range (i.e. not before the first requested base); add this datapoint
		if !(chromIndex < filterStartChromIndex || (chromIndex == filterStartChromIndex && endBase < filterStartBase)) {
			decodedData = append(decodedData, BigWigData{
				Chr:   chrom,
				Start: startBase,
				End:   endBase,
				Value: value,
			})
		}

		// For Fixed Step format, increment the start base after processing
		if dataType != 1 && dataType != 2 {
			startBase += itemStep
		}
	}

	return decodedData, nil
}
