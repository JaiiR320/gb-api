package bigdata

const (
	RPTREE_HEADER_SIZE = 48         // R+ tree header size
	RPTREE_NODE_LEAF   = 1          // R+ tree leaf node type
	RPTREE_NODE_CHILD  = 0          // R+ tree non-leaf node type
	CHROM_TREE_MAGIC   = 0x78CA8C91 // Chrom Tree Magic Number
	IDX_MAGIC          = 0x2468ACE0 // R+ tree index magic number
)

type ChromTree struct {
	BlockSize int32            `json:"blockSize"`
	KeySize   int32            `json:"keySize"`
	ValSize   int32            `json:"valSize"`
	ItemCount uint64           `json:"itemCount"`
	Reserved  uint64           `json:"reserved"`
	ChromToID map[string]int32 `json:"chromToId"`
	ChromSize map[string]int32 `json:"chromSize"`
	IDToChrom map[int32]string `json:"idToChrom"`
}

// RPTreeHeader represents the R+ tree header
type RPTreeHeader struct {
	Magic         uint32
	BlockSize     uint32
	ItemCount     uint64
	StartChromIx  uint32
	StartBase     uint32
	EndChromIx    uint32
	EndBase       uint32
	EndFileOffset uint64
	ItemsPerSlot  uint32
	Reserved      uint32
}

// RPLeafNode represents a leaf node in the R+ tree
type RPLeafNode struct {
	StartChromIx uint32
	StartBase    uint32
	EndChromIx   uint32
	EndBase      uint32
	DataOffset   uint64
	DataSize     uint64
}

// RPChildNode represents a child node in the R+ tree
type RPChildNode struct {
	StartChromIx uint32
	StartBase    uint32
	EndChromIx   uint32
	EndBase      uint32
	ChildOffset  uint64
}
