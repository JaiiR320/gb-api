# Sort
{
  grep '^#' unsorted.gtf
  grep -v '^#' unsorted.gtf | sort -k1,1 -k4,4n
} > sorted.gtf

# Compress with BGZF
bgzip sorted.gtf   # produces sorted.gtf.gz

# Index for region queries
tabix -p gff sorted.gtf.gz
