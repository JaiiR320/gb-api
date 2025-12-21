package bigwig

const (
	RPTREE_HEADER_SIZE = 48         // R+ tree header size
	RPTREE_NODE_LEAF   = 1          // R+ tree leaf node type
	RPTREE_NODE_CHILD  = 0          // R+ tree non-leaf node type
	CHROM_TREE_MAGIC   = 0x78CA8C91 // Chrom Tree Magic Number
	IDX_MAGIC          = 0x2468ACE0 // R+ tree index magic number
)
