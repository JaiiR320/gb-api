package bigwig

import "fmt"

// PrintHeader prints the BigWig header information
func (b *BigWig) PrintHeader() {
	fmt.Printf("BigWig Header:\n")
	fmt.Printf("  bwVersion: %d\n", b.Header.BwVersion)
	fmt.Printf("  nZoomLevels: %d\n", b.Header.NZoomLevels)
	fmt.Printf("  chromTreeOffset: %d\n", b.Header.ChromTreeOffset)
	fmt.Printf("  fullDataOffset: %d\n", b.Header.FullDataOffset)
	fmt.Printf("  fullIndexOffset: %d\n", b.Header.FullIndexOffset)
	fmt.Printf("  fieldCount: %d\n", b.Header.FieldCount)
	fmt.Printf("  definedFieldCount: %d\n", b.Header.DefinedFieldCount)
	fmt.Printf("  autoSqlOffset: %d\n", b.Header.AutoSqlOffset)
	fmt.Printf("  totalSummaryOffset: %d\n", b.Header.TotalSummaryOffset)
	fmt.Printf("  uncompressBuffSize: %d\n", b.Header.UncompressBuffSize)
	fmt.Printf("  reserved: %d\n", b.Header.Reserved)
}

// Print prints all BigWig metadata including header, zoom levels, and chromosome tree
func (b *BigWig) Print() {
	b.PrintHeader()

	fmt.Printf("\nZoom Levels:\n")
	for _, zl := range b.ZoomLevels {
		fmt.Printf("  Zoom Level %d:\n", zl.Index)
		fmt.Printf("    reductionLevel: %d\n", zl.ReductionLevel)
		fmt.Printf("    reserved: %d\n", zl.Reserved)
		fmt.Printf("    dataOffset: %d\n", zl.DataOffset)
		fmt.Printf("    indexOffset: %d\n", zl.IndexOffset)
	}

	fmt.Printf("\nTotal Summary:\n")
	fmt.Printf("  basesCovered: %d\n", b.TotalSummary.BasesCovered)
	fmt.Printf("  minVal: %f\n", b.TotalSummary.MinVal)
	fmt.Printf("  maxVal: %f\n", b.TotalSummary.MaxVal)
	fmt.Printf("  sumData: %f\n", b.TotalSummary.SumData)
	fmt.Printf("  sumSquares: %f\n", b.TotalSummary.SumSquares)

	fmt.Printf("\nChromosome Tree:\n")
	for chrom, id := range b.ChromTree.ChromToID {
		size := b.ChromTree.ChromSize[chrom]
		fmt.Printf("  Chromosome: %s, ID: %d, Size: %d\n", chrom, id, size)
	}
}
