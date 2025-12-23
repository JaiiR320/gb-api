package bigdata

import (
	"bytes"
	"encoding/binary"
	"gb-api/utils"
)

// LoadLeafNodesForRPNode recursively loads leaf nodes from the R+ tree
func LoadLeafNodesForRPNode(url string, byteOrder binary.ByteOrder, nodeOffset uint64, startChromIx int32, startBase int32,
	endChromIx int32, endBase int32) ([]RPLeafNode, error) {

	// Fetch header + node data in single request (4KB prefetch buffer)
	data, err := RequestBytes(url, int(nodeOffset), RPTREE_NODE_PREFETCH_SIZE)
	if err != nil {
		return nil, err
	}

	// Parse header from first 4 bytes
	p := utils.NewParser(bytes.NewReader(data[:4]), byteOrder)
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

	// Calculate required data size based on node type
	var itemSize int
	if isLeaf == RPTREE_NODE_LEAF {
		itemSize = 32 // Leaf items: 4×uint32 + 2×uint64 = 32 bytes
	} else {
		itemSize = 24 // Child items: 4×uint32 + 1×uint64 = 24 bytes
	}
	requiredDataSize := int(count) * itemSize

	// Check if all data is already in prefetch buffer
	var nodeData []byte
	if 4+requiredDataSize <= RPTREE_NODE_PREFETCH_SIZE {
		// Common case: all data in buffer, no additional fetch needed
		nodeData = data[4 : 4+requiredDataSize]
	} else {
		// Rare case: node exceeds 4KB, fetch remaining data
		nodeData, err = RequestBytes(url, int(nodeOffset)+4, requiredDataSize)
		if err != nil {
			return nil, err
		}
	}

	// Create parser for node data
	p = utils.NewParser(bytes.NewReader(nodeData), byteOrder)

	leafNodes := []RPLeafNode{}

	if isLeaf == RPTREE_NODE_LEAF {
		// This is a leaf node - read leaf items directly from nodeData (already in memory)
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
		// This is a child node - parse all children and filter for overlapping ones
		overlappingChildren := make([]RPChildNode, 0, count)
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

			// Only process children that overlap with query range
			if overlaps(child.StartChromIx, child.StartBase, child.EndChromIx, child.EndBase,
				uint32(startChromIx), uint32(startBase), uint32(endChromIx), uint32(endBase)) {
				overlappingChildren = append(overlappingChildren, child)
			}
		}

		// Process overlapping children in parallel
		type childResult struct {
			leaves []RPLeafNode
			err    error
		}

		resultsChan := make(chan childResult, len(overlappingChildren))

		// Spawn goroutine for each overlapping child
		for _, child := range overlappingChildren {
			childCopy := child // Capture for goroutine closure
			go func() {
				childLeaves, err := LoadLeafNodesForRPNode(
					url, byteOrder, childCopy.ChildOffset,
					startChromIx, startBase, endChromIx, endBase,
				)
				resultsChan <- childResult{leaves: childLeaves, err: err}
			}()
		}

		// Collect results from all goroutines
		for range overlappingChildren {
			result := <-resultsChan
			if result.err != nil {
				return nil, result.err
			}
			leafNodes = append(leafNodes, result.leaves...)
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
