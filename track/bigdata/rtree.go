package bigdata

import (
	"bytes"
	"encoding/binary"
	"gb-api/utils"
)

// LoadLeafNodesForRPNode recursively loads leaf nodes from the R+ tree
func LoadLeafNodesForRPNode(url string, byteOrder binary.ByteOrder, nodeOffset uint64, startChromIx int32, startBase int32,
	endChromIx int32, endBase int32) ([]RPLeafNode, error) {

	data, err := RequestBytes(url, int(nodeOffset), 4)
	if err != nil {
		return nil, err
	}

	p := utils.NewParser(bytes.NewReader(data), byteOrder)
	isLeaf, err := p.GetUInt8()
	if err != nil {
		return nil, err
	}

	// Skip reserved byte
	_, err = p.GetUInt8()
	if err != nil {
		return nil, err
	}

	count, err := p.GetUInt16()
	if err != nil {
		return nil, err
	}

	leafNodes := []RPLeafNode{}

	if isLeaf == RPTREE_NODE_LEAF {
		// This is a leaf node - read leaf items
		itemSize := 32 // Each leaf item is 32 bytes
		data, err := RequestBytes(url, int(nodeOffset)+4, int(count)*itemSize)
		if err != nil {
			return nil, err
		}

		p := utils.NewParser(bytes.NewReader(data), byteOrder)

		for range count {
			var leaf RPLeafNode
			err := p.ReadMultiple(
				&leaf.StartChromIx,
				&leaf.StartBase,
				&leaf.EndChromIx,
				&leaf.EndBase,
				&leaf.DataOffset,
				&leaf.DataSize,
			)
			if err != nil {
				return nil, err
			}

			// Check if this leaf overlaps with our query range
			if overlaps(leaf.StartChromIx, leaf.StartBase, leaf.EndChromIx, leaf.EndBase,
				uint32(startChromIx), uint32(startBase), uint32(endChromIx), uint32(endBase)) {
				leafNodes = append(leafNodes, leaf)
			}
		}
	} else {
		// This is a child node - recursively process children
		itemSize := 24 // Each child item is 24 bytes
		data, err := RequestBytes(url, int(nodeOffset)+4, int(count)*itemSize)
		if err != nil {
			return nil, err
		}

		p := utils.NewParser(bytes.NewReader(data), byteOrder)

		for range count {
			var child RPChildNode
			err := p.ReadMultiple(
				&child.StartChromIx,
				&child.StartBase,
				&child.EndChromIx,
				&child.EndBase,
				&child.ChildOffset,
			)
			if err != nil {
				return nil, err
			}

			// Check if this child overlaps with our query range
			if overlaps(child.StartChromIx, child.StartBase, child.EndChromIx, child.EndBase,
				uint32(startChromIx), uint32(startBase), uint32(endChromIx), uint32(endBase)) {
				childLeaves, err := LoadLeafNodesForRPNode(url, byteOrder, child.ChildOffset, startChromIx, startBase, endChromIx, endBase)
				if err != nil {
					return nil, err
				}
				leafNodes = append(leafNodes, childLeaves...)
			}
		}
	}

	return leafNodes, nil
}

// overlaps checks if two genomic ranges overlap
func overlaps(aStartChrom, aStartBase, aEndChrom, aEndBase, bStartChrom, bStartBase, bEndChrom, bEndBase uint32) bool {
	// If a ends before b starts, no overlap
	if aEndChrom < bStartChrom || (aEndChrom == bStartChrom && aEndBase <= bStartBase) {
		return false
	}
	// If a starts after b ends, no overlap
	if aStartChrom > bEndChrom || (aStartChrom == bEndChrom && aStartBase >= bEndBase) {
		return false
	}
	return true
}
