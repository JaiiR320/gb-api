package bigwig

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"gb-api/track/common"
	"gb-api/utils"
	"io"
)

// ParseBedFunction is a function that parses the rest of a BED record
type ParseBedFunction[T any] func(chrom string, startBase, endBase int32, rest string) T

// DecodeBedData decodes BED data from a binary buffer with filtering
func (b *BigWig) DecodeBedData(data []byte, filterStartChromIndex int32, filterStartBase int32,
	filterEndChromIndex int32, filterEndBase int32, chromDict []string, restParser ParseBedFunction[any]) ([]any, error) {

	decodedData := []any{}
	binaryParser := utils.NewParser(bytes.NewReader(data), b.ByteOrder)

	for {
		// Check if enough bytes remain (simplified check)
		chromIndex, err := binaryParser.GetInt32()
		if err != nil {
			break
		}

		startBase, err := binaryParser.GetInt32()
		if err != nil {
			break
		}

		endBase, err := binaryParser.GetInt32()
		if err != nil {
			break
		}

		rest, err := binaryParser.GetString(0) // 0 means read until null terminator
		if err != nil {
			break
		}

		// Get chromosome name
		if chromIndex < 0 || int(chromIndex) >= len(chromDict) {
			continue
		}
		chrom := chromDict[chromIndex]

		// Filter: skip entries before the start
		if chromIndex < filterStartChromIndex || (chromIndex == filterStartChromIndex && endBase < filterStartBase) {
			continue
		}

		// Filter: stop processing entries after the end
		if chromIndex > filterEndChromIndex || (chromIndex == filterEndChromIndex && startBase >= filterEndBase) {
			break
		}

		// Parse and add entry
		entry := restParser(chrom, startBase, endBase, rest)
		decodedData = append(decodedData, entry)
	}

	return decodedData, nil
}

// DecodeWigData decodes BigWig data from a binary buffer
func (b *BigWig) DecodeWigData(data []byte, filterStartChromIndex int32, filterStartBase int32,
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

// decompressData decompresses zlib-compressed data if needed
func decompressData(data []byte, needsDecompression bool) ([]byte, error) {
	if !needsDecompression {
		return data, nil
	}

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return decompressed, nil
}

// ReadBigWigData reads BigWig data for a genomic region
func (b *BigWig) ReadBigWigData(startChrom string, startBase int32, endChrom string, endBase int32) ([]BigWigData, error) {
	// Get chromosome indices
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
	headerData, err := common.RequestBytes(b.URL, int(treeOffset), common.RPTREE_HEADER_SIZE)
	if err != nil {
		return nil, err
	}

	p := utils.NewParser(bytes.NewReader(headerData), b.ByteOrder)
	magic, err := p.GetUInt32()
	if err != nil {
		return nil, err
	}

	if magic != common.IDX_MAGIC {
		return nil, fmt.Errorf("R+ tree not found at offset %d", treeOffset)
	}

	// Load leaf nodes from R+ tree
	rootNodeOffset := treeOffset + common.RPTREE_HEADER_SIZE
	leafNodes, err := loadLeafNodesForRPNode(b, rootNodeOffset, startChromIndex, startBase, endChromIndex, endBase)
	if err != nil {
		return nil, err
	}

	// Iterate through leaf nodes and decode data
	allData := []BigWigData{}
	for _, leafNode := range leafNodes {
		leafData, err := common.RequestBytes(b.URL, int(leafNode.DataOffset), int(leafNode.DataSize))
		if err != nil {
			return nil, err
		}

		// Decompress if necessary
		leafData, err = decompressData(leafData, b.Header.UncompressBuffSize > 0)
		if err != nil {
			return nil, err
		}

		// Decode the data
		decodedData, err := b.DecodeWigData(leafData, startChromIndex, startBase, endChromIndex, endBase)
		if err != nil {
			return nil, err
		}

		allData = append(allData, decodedData...)
	}

	return allData, nil
}
