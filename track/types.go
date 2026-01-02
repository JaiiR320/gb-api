package track

type GenomicRange struct {
	Chrom string `json:"chr"`
	Start string `json:"start"`
	End   string `json:"end"`
}
