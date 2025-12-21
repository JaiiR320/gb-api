package bigwig

const (
	TWOBIT_MAGIC_LTH    = 0x1A412743 // BigWig Magic High to Low
	TWOBIT_MAGIC_HTL    = 0x4327411A // BigWig Magic Low to High
	BIGWIG_MAGIC_LTH    = 0x888FFC26 // BigWig Magic Low to High
	BIGWIG_MAGIC_HTL    = 0x26FC8F88 // BigWig Magic High to Low
	BIGBED_MAGIC_LTH    = 0x8789F2EB // BigBed Magic Low to High
	BIGBED_MAGIC_HTL    = 0xEBF28987 // BigBed Magic High to Low
	BBFILE_HEADER_SIZE  = 64         // Common header size
	DEFAULT_BUFFER_SIZE = 512000     // Default buffer size for data loading
)
