package bigdata

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"gb-api/utils"
)

// LoadHeader loads and parses the BigBed file header
func (b *BigData) LoadHeader() error {
	data, err := RequestBytes(b.URL, 0, BBFILE_HEADER_SIZE)
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
	if magic != b.LTH {
		byteOrder = binary.BigEndian
		p = utils.NewParser(bytes.NewReader(data), byteOrder)

		magic, err = p.GetUInt32()
		if err != nil {
			return err
		}
		if magic != b.HTL {
			return fmt.Errorf("invalid file magic number: 0x%08X", magic)
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
func (b *BigData) LoadMetaData() error {
	data, err := RequestBytes(b.URL, 64, int(b.Header.FullDataOffset)-64+5)
	if err != nil {
		return err
	}

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
		_, err := p.SetPosition(FileOffsetToBufferOffset(b.Header.AutoSqlOffset), 0)
		if err != nil {
			return err
		}
		autosql, err = p.GetString(0)
		if err != nil {
			return err
		}
	}

	b.AutoSql = autosql

	var totalSummary TotalSummary
	if b.Header.TotalSummaryOffset != 0 {
		_, err := p.SetPosition(FileOffsetToBufferOffset(b.Header.TotalSummaryOffset), 0)
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
	_, err = p.SetPosition(FileOffsetToBufferOffset(b.Header.ChromTreeOffset), 0)
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

	err = BuildChromTree(&chromTree, p, nil)
	if err != nil {
		return err
	}
	b.ChromTree = chromTree

	treeOffset := b.Header.FullIndexOffset
	headerData, err := RequestBytes(b.URL, int(treeOffset), RPTREE_HEADER_SIZE)
	if err != nil {
		return err
	}

	p = utils.NewParser(bytes.NewReader(headerData), b.ByteOrder)
	magic, err = p.GetUInt32()
	if err != nil {
		return err
	}

	if magic != IDX_MAGIC {
		return fmt.Errorf("R+ tree not found at offset %d", treeOffset)
	}

	return nil
}
