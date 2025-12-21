package bigbed

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"gb-api/track/bigwig"
	"gb-api/utils"
	"io"
)

const (
	RPTREE_HEADER_SIZE = 48
	IDX_MAGIC          = 0x2468ACE0
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

// ReadBigBedData reads BigBed data for a genomic region
func (b *BigBed) ReadBigBedData(startChrom string, startBase int32, endChrom string, endBase int32) ([]BigBedData, error) {
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
	headerData, err := bigwig.RequestBytes(b.URL, int(treeOffset), RPTREE_HEADER_SIZE)
	if err != nil {
		return nil, err
	}

	p := utils.NewParser(bytes.NewReader(headerData), b.ByteOrder)
	magic, err := p.GetUInt32()
	if err != nil {
		return nil, err
	}

	if magic != IDX_MAGIC {
		return nil, fmt.Errorf("R+ tree not found at offset %d", treeOffset)
	}

	// Load leaf nodes from R+ tree
	rootNodeOffset := treeOffset + RPTREE_HEADER_SIZE
	leafNodes, err := bigwig.LoadLeafNodesForRPNode(b.URL, b.ByteOrder, rootNodeOffset, startChromIndex, startBase, endChromIndex, endBase)
	if err != nil {
		return nil, err
	}

	// Iterate through leaf nodes and decode data
	allData := []BigBedData{}
	for _, leafNode := range leafNodes {
		leafData, err := bigwig.RequestBytes(b.URL, int(leafNode.DataOffset), int(leafNode.DataSize))
		if err != nil {
			return nil, err
		}

		// Decompress if necessary
		leafData, err = decompressData(leafData, b.Header.UncompressBuffSize > 0)
		if err != nil {
			return nil, err
		}

		// Decode the data
		decodedData, err := b.DecodeBedData(leafData, startChromIndex, startBase, endChromIndex, endBase)
		if err != nil {
			return nil, err
		}

		allData = append(allData, decodedData...)
	}

	return allData, nil
}
