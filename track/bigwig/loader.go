package bigwig

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"gb-api/track/common"
	"gb-api/utils"
)

// loadLeafNodesForRPNode is the internal version used by BigWig
func loadLeafNodesForRPNode(b *BigWig, nodeOffset uint64, startChromIx int32, startBase int32,
	endChromIx int32, endBase int32) ([]common.RPLeafNode, error) {
	return common.LoadLeafNodesForRPNode(b.URL, b.ByteOrder, nodeOffset, startChromIx, startBase, endChromIx, endBase)
}

// LoadHeader loads and parses the BigWig file header
func (b *BigWig) LoadHeader() error {
	data, err := common.RequestBytes(b.URL, 0, common.BBFILE_HEADER_SIZE)
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
	if magic != BIGWIG_MAGIC_LTH {
		byteOrder = binary.BigEndian
		p = utils.NewParser(bytes.NewReader(data), byteOrder)

		magic, err = p.GetUInt32()
		if err != nil {
			return err
		}
		if magic != BIGWIG_MAGIC_HTL {
			return fmt.Errorf("invalid BigWig magic number: 0x%08X", magic)
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

// LoadMetaData loads BigWig metadata including zoom levels, autoSql, total summary, and chromosome tree
func (b *BigWig) LoadMetaData(data []byte) error {
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
		_, err := p.SetPosition(common.FileOffsetToBufferOffset(b.Header.AutoSqlOffset), 0)
		if err != nil {
			return err
		}
		autosql, err = p.GetString(0)
		if err != nil {
			return err
		}
	}

	b.AutoSql = autosql

	var totalSummary BWTotalSummary
	if b.Header.TotalSummaryOffset != 0 {
		_, err := p.SetPosition(common.FileOffsetToBufferOffset(b.Header.TotalSummaryOffset), 0)
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

	var chromTree common.ChromTree
	_, err := p.SetPosition(common.FileOffsetToBufferOffset(b.Header.ChromTreeOffset), 0)
	if err != nil {
		return err
	}
	magic, err := p.GetUInt32()
	if err != nil {
		return err
	}
	if magic != common.CHROM_TREE_MAGIC {
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

	err = common.BuildChromTree(&chromTree, p, nil)
	if err != nil {
		return err
	}
	b.ChromTree = chromTree

	return nil
}
