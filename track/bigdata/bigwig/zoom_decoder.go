package bigwig

import (
	"bytes"
	"gb-api/track/bigdata"
	"gb-api/utils"
)

// ZoomRecord represents a single zoom summary record (32 bytes)
type ZoomRecord struct {
	ChromId    int32
	Start      int32
	End        int32
	ValidCount uint32
	MinVal     float32
	MaxVal     float32
	SumData    float32
	SumSquares float32
}

// decodeZoomData decodes zoom level summary data into BigWigData points.
// Each zoom record is converted to a single BigWigData point using mean value.
// Per requirements: value = sumData / validCount
func decodeZoomData(b *bigdata.BigData, data []byte, filterStartChromIndex int32,
	filterStartBase int32, filterEndChromIndex int32, filterEndBase int32) ([]BigWigData, error) {

	decodedData := []BigWigData{}
	binaryParser := utils.NewParser(bytes.NewReader(data), b.ByteOrder)

	// Each zoom record is exactly 32 bytes
	const ZOOM_RECORD_SIZE = 32
	recordCount := len(data) / ZOOM_RECORD_SIZE

	for i := 0; i < recordCount; i++ {
		var record ZoomRecord

		err := binaryParser.ReadMultiple(
			&record.ChromId,
			&record.Start,
			&record.End,
			&record.ValidCount,
			&record.MinVal,
			&record.MaxVal,
			&record.SumData,
			&record.SumSquares,
		)
		if err != nil {
			return nil, err
		}

		// Filter by chromosome and position
		if record.ChromId < filterStartChromIndex || record.ChromId > filterEndChromIndex {
			continue
		}

		// Check if past the end of the range
		if record.ChromId > filterEndChromIndex ||
			(record.ChromId == filterEndChromIndex && record.Start >= filterEndBase) {
			break
		}

		// Check if before the start of the range
		if record.ChromId < filterStartChromIndex ||
			(record.ChromId == filterStartChromIndex && record.End < filterStartBase) {
			continue
		}

		// Calculate mean value (per requirements)
		var value float32
		if record.ValidCount > 0 {
			value = record.SumData / float32(record.ValidCount)
		} else {
			// Edge case: no valid data in this bin, use 0
			value = 0
		}

		chrom := b.ChromTree.IDToChrom[record.ChromId]

		decodedData = append(decodedData, BigWigData{
			Chr:   chrom,
			Start: record.Start,
			End:   record.End,
			Value: value,
		})
	}

	return decodedData, nil
}
