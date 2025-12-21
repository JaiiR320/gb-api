package bigdata

import (
	"bytes"
	"fmt"
	"gb-api/utils"
)

// Decoder function signature
type DataDecoder[T any] func(b *BigData, data []byte, startChromIdx, startBase, endChromIdx, endBase int32) ([]T, error)

func ReadData[T any](
	b *BigData,
	chrom string, start int32, end int32,
	decoder DataDecoder[T],
) ([]T, error) {
	startChromIndex, ok := b.ChromTree.ChromToID[chrom]
	if !ok {
		return nil, fmt.Errorf("chromosome %s not found", chrom)
	}
	endChromIndex := startChromIndex // same chrom

	// Read R+ tree header
	treeOffset := b.Header.FullIndexOffset
	headerData, err := RequestBytes(b.URL, int(treeOffset), RPTREE_HEADER_SIZE)
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
	leafNodes, err := LoadLeafNodesForRPNode(b.URL, b.ByteOrder, rootNodeOffset, startChromIndex, start, endChromIndex, end)
	if err != nil {
		return nil, err
	}

	// Iterate through leaf nodes and decode data
	allData := []T{}
	for _, leafNode := range leafNodes {
		leafData, err := RequestBytes(b.URL, int(leafNode.DataOffset), int(leafNode.DataSize))
		if err != nil {
			return nil, err
		}

		leafData, err = DecompressData(leafData, b.Header.UncompressBuffSize > 0)
		if err != nil {
			return nil, err
		}

		decodedData, err := decoder(b, leafData, startChromIndex, start, endChromIndex, end)
		if err != nil {
			return nil, err
		}

		allData = append(allData, decodedData...)
	}

	return allData, nil
}
