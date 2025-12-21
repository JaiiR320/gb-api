package bigdata

import (
	"gb-api/utils"
)

// buildChromTree recursively builds the chromosome B+ tree
func BuildChromTree(chromTree *ChromTree, binaryParser *utils.Parser, offset *int64) error {
	if offset != nil {
		_, err := binaryParser.SetPosition(*offset, 0)
		if err != nil {
			return err
		}
	}

	nodeType, err := binaryParser.GetUInt8()
	if err != nil {
		return err
	}

	// Skip reserved byte
	_, err = binaryParser.GetUInt8()
	if err != nil {
		return err
	}

	count, err := binaryParser.GetUInt16()
	if err != nil {
		return err
	}

	// If the node is a leaf (type == 1)
	if nodeType == 1 {
		for range count {
			key, err := binaryParser.GetFixedLengthTrimmedString(int(chromTree.KeySize))
			if err != nil {
				return err
			}

			chromID, err := binaryParser.GetInt32()
			if err != nil {
				return err
			}

			chromSize, err := binaryParser.GetInt32()
			if err != nil {
				return err
			}

			chromTree.ChromToID[key] = chromID
			chromTree.IDToChrom[chromID] = key
			chromTree.ChromSize[key] = chromSize
		}
	} else {
		// Internal node
		for range count {
			_, err := binaryParser.GetFixedLengthTrimmedString(int(chromTree.KeySize))
			if err != nil {
				return err
			}

			childOffset, err := binaryParser.GetUInt64()
			if err != nil {
				return err
			}

			bufferOffset := FileOffsetToBufferOffset(childOffset)

			// Save current position
			currPos, err := binaryParser.SetPosition(0, 1) // Get current position
			if err != nil {
				return err
			}

			// Recursively build child node
			err = BuildChromTree(chromTree, binaryParser, &bufferOffset)
			if err != nil {
				return err
			}

			// Restore position
			_, err = binaryParser.SetPosition(currPos, 0)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// FileOffsetToBufferOffset converts a file offset to a buffer offset
// by subtracting the header size
func FileOffsetToBufferOffset(offset uint64) int64 {
	return int64(offset - uint64(BBFILE_HEADER_SIZE))
}
