package bigbed

import (
	"bytes"
	"fmt"
	"gb-api/track/bigdata"
	"gb-api/utils"
)

// ParseBedFunction is a function that parses the rest of a BED record
type ParseBedFunction[T any] func(chrom string, startBase, endBase int32, rest string) T

// DecodeBedData decodes BED data from a binary buffer with filtering
func (b *BigBed) DecodeBedData(data []byte, filterStartChromIndex int32, filterStartBase int32,
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

// ReadBigBedData reads BigBed data for a genomic region
func (b *BigBed) ReadBigBedData(chrom string, start int32, end int32) ([]BigBedData, error) {
	// Get chromosome indices
	startChrom := chrom
	endChrom := chrom
	startChromIndex, ok := b.ChromTree.ChromToID[startChrom]
	if !ok {
		return nil, fmt.Errorf("chromosome %s not found", startChrom)
	}

	endChromIndex, ok := b.ChromTree.ChromToID[endChrom]
	if !ok {
		return nil, fmt.Errorf("chromosome %s not found", endChrom)
	}

	// Read R+ tree header
	treeOffset := b.Header.FullIndexOffset
	headerData, err := bigdata.RequestBytes(b.URL, int(treeOffset), bigdata.RPTREE_HEADER_SIZE)
	if err != nil {
		return nil, err
	}

	p := utils.NewParser(bytes.NewReader(headerData), b.ByteOrder)
	magic, err := p.GetUInt32()
	if err != nil {
		return nil, err
	}

	if magic != bigdata.IDX_MAGIC {
		return nil, fmt.Errorf("R+ tree not found at offset %d", treeOffset)
	}

	// Load leaf nodes from R+ tree
	rootNodeOffset := treeOffset + bigdata.RPTREE_HEADER_SIZE
	leafNodes, err := bigdata.LoadLeafNodesForRPNode(b.URL, b.ByteOrder, rootNodeOffset, startChromIndex, start, endChromIndex, end)
	if err != nil {
		return nil, err
	}

	// Iterate through leaf nodes and decode data
	allData := []BigBedData{}
	for _, leafNode := range leafNodes {
		leafData, err := bigdata.RequestBytes(b.URL, int(leafNode.DataOffset), int(leafNode.DataSize))
		if err != nil {
			return nil, err
		}

		// Decompress if necessary
		leafData, err = bigdata.DecompressData(leafData, b.Header.UncompressBuffSize > 0)
		if err != nil {
			return nil, err
		}

		// Decode the data
		decodedData, err := b.DecodeBedData(leafData, startChromIndex, start, endChromIndex, end)
		if err != nil {
			return nil, err
		}

		allData = append(allData, decodedData...)
	}

	return allData, nil
}
