package bigdata

import (
	"fmt"
)

// Decoder function signature
type DataDecoder[T any] func(b *BigData, data []byte, startChromIdx, startBase, endChromIdx, endBase int32) ([]T, error)

// ReadDataWithZoom reads data from either full resolution or a zoom level
func ReadDataWithZoom[T any](
	b *BigData,
	chrom string, start int32, end int32,
	decoder DataDecoder[T],
	zoomLevelIndex int, // -1 for full resolution, >=0 for zoom level
) ([]T, error) {
	startChromIndex, ok := b.ChromTree.ChromToID[chrom]
	if !ok {
		return nil, fmt.Errorf("chromosome %s not found", chrom)
	}
	endChromIndex := startChromIndex // same chrom

	// Determine which R+ tree to use
	var treeOffset uint64
	if zoomLevelIndex >= 0 && zoomLevelIndex < len(b.ZoomLevels) {
		// Use zoom level's R+ tree
		treeOffset = b.ZoomLevels[zoomLevelIndex].IndexOffset
	} else {
		// Use full resolution R+ tree
		treeOffset = b.Header.FullIndexOffset
	}

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

// ReadData is maintained for backward compatibility
func ReadData[T any](
	b *BigData,
	chrom string, start int32, end int32,
	decoder DataDecoder[T],
) ([]T, error) {
	return ReadDataWithZoom(b, chrom, start, end, decoder, -1)
}
