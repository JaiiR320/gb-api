package bigbed

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gb-api/track/common"
	"gb-api/utils"
)

const (
	BIGBED_MAGIC_LTH   = 0x8789F2EB // BigBed Magic Low to High
	BIGBED_MAGIC_HTL   = 0xEBF28987 // BigBed Magic High to Low
	BBFILE_HEADER_SIZE = 64         // Common header size
	CHROM_TREE_MAGIC   = 0x78CA8C91 // Chrom Tree Magic Number
)

// fileOffsetToBufferOffset converts a file offset to a buffer offset
// by subtracting the header size
func fileOffsetToBufferOffset(offset uint64) int64 {
	return int64(offset - uint64(BBFILE_HEADER_SIZE))
}

// LoadHeader loads and parses the BigBed file header
func (b *BigBed) LoadHeader() error {
	data, err := common.RequestBytes(b.URL, 0, BBFILE_HEADER_SIZE)
	if err != nil {
		return err
	}

	// Try little endian first
	var byteOrder binary.ByteOrder = binary.LittleEndian
	p := utils.NewParser(bytes.NewReader(data), byteOrder)

	magic, err := p.GetUInt32()
	if err != nil {
		return err
	}

	// If magic doesn't match, try big endian
	if magic != BIGBED_MAGIC_LTH {
		byteOrder = binary.BigEndian
		p = utils.NewParser(bytes.NewReader(data), byteOrder)

		magic, err = p.GetUInt32()
		if err != nil {
			return err
		}
		if magic != BIGBED_MAGIC_HTL {
			return fmt.Errorf("invalid BigBed magic number: 0x%08X", magic)
		}
	}

	var header Header
	if err := p.ReadStruct(&header); err != nil {
		return err
	}

	b.Header = header
	b.ByteOrder = byteOrder

	return nil
}

// LoadMetaData loads BigBed metadata including zoom levels, autoSql, total summary, and chromosome tree
func (b *BigBed) LoadMetaData(data []byte) error {
	p := utils.NewParser(bytes.NewReader(data), b.ByteOrder)

	b.ZoomLevels = make([]ZoomLevelHeader, b.Header.NZoomLevels)

	for i := 1; i <= int(b.Header.NZoomLevels); i++ {
		zoomNumber := int(b.Header.NZoomLevels) - i

		b.ZoomLevels[zoomNumber].Index = zoomNumber

		if err := p.ReadMultiple(
			&b.ZoomLevels[zoomNumber].ReductionLevel,
			&b.ZoomLevels[zoomNumber].Reserved,
			&b.ZoomLevels[zoomNumber].DataOffset,
			&b.ZoomLevels[zoomNumber].IndexOffset,
		); err != nil {
			return err
		}
	}

	var autosql string
	if b.Header.AutoSqlOffset != 0 {
		_, err := p.SetPosition(fileOffsetToBufferOffset(b.Header.AutoSqlOffset), 0)
		if err != nil {
			return err
		}
		autosql, err = p.GetString(0)
		if err != nil {
			return err
		}
	}

	b.AutoSql = autosql

	var totalSummary BBTotalSummary
	if b.Header.TotalSummaryOffset != 0 {
		_, err := p.SetPosition(fileOffsetToBufferOffset(b.Header.TotalSummaryOffset), 0)
		if err != nil {
			return err
		}

		if err := p.ReadMultiple(
			&totalSummary.BasesCovered,
			&totalSummary.MinVal,
			&totalSummary.MaxVal,
			&totalSummary.SumData,
			&totalSummary.SumSquares,
		); err != nil {
			return err
		}
	}
	b.TotalSummary = totalSummary

	var chromTree ChromTree
	_, err := p.SetPosition(fileOffsetToBufferOffset(b.Header.ChromTreeOffset), 0)
	if err != nil {
		return err
	}
	magic, err := p.GetUInt32()
	if err != nil {
		return err
	}
	if magic != CHROM_TREE_MAGIC {
		return fmt.Errorf("chromosome B+ tree not found at offset %d", b.Header.ChromTreeOffset)
	}

	if err := p.ReadMultiple(
		&chromTree.BlockSize,
		&chromTree.KeySize,
		&chromTree.ValSize,
		&chromTree.ItemCount,
		&chromTree.Reserved,
	); err != nil {
		return err
	}

	chromTree.ChromToID = make(map[string]int32)
	chromTree.ChromSize = make(map[string]int32)
	chromTree.IDToChrom = make(map[int32]string)

	err = buildChromTree(&chromTree, p, nil)
	if err != nil {
		return err
	}
	b.ChromTree = chromTree

	return nil
}

// buildChromTree recursively builds the chromosome tree
func buildChromTree(tree *ChromTree, p *utils.Parser, parent *chromTreeNode) error {
	node, err := readChromTreeNode(p, tree)
	if err != nil {
		return err
	}

	if node.isLeaf {
		for _, key := range node.keys {
			tree.ChromToID[key.name] = key.chromID
			tree.ChromSize[key.name] = key.chromSize
			tree.IDToChrom[key.chromID] = key.name
		}
	} else {
		for _, child := range node.children {
			_, err := p.SetPosition(int64(child), 0)
			if err != nil {
				return err
			}
			if err := buildChromTree(tree, p, node); err != nil {
				return err
			}
		}
	}

	return nil
}

type chromTreeNode struct {
	isLeaf   bool
	keys     []chromKey
	children []uint64
}

type chromKey struct {
	name      string
	chromID   int32
	chromSize int32
}

func readChromTreeNode(p *utils.Parser, tree *ChromTree) (*chromTreeNode, error) {
	node := &chromTreeNode{}

	isLeaf, err := p.GetUInt8()
	if err != nil {
		return nil, err
	}
	node.isLeaf = isLeaf == 1

	// Skip reserved byte
	_, err = p.GetUInt8()
	if err != nil {
		return nil, err
	}

	count, err := p.GetUInt16()
	if err != nil {
		return nil, err
	}

	if node.isLeaf {
		node.keys = make([]chromKey, count)
		for i := range count {
			key, err := p.GetFixedLengthString(int(tree.KeySize))
			if err != nil {
				return nil, err
			}

			chromID, err := p.GetInt32()
			if err != nil {
				return nil, err
			}

			chromSize, err := p.GetInt32()
			if err != nil {
				return nil, err
			}

			node.keys[i] = chromKey{
				name:      key,
				chromID:   chromID,
				chromSize: chromSize,
			}
		}
	} else {
		node.children = make([]uint64, count)
		for i := range count {
			// Skip key
			_, err := p.GetFixedLengthString(int(tree.KeySize))
			if err != nil {
				return nil, err
			}

			childOffset, err := p.GetUInt64()
			if err != nil {
				return nil, err
			}

			node.children[i] = childOffset
		}
	}

	return node, nil
}
