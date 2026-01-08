package transcript

import (
	"strconv"
)

func GetTranscripts(chrom string, start int, end int) ([]Gene, error) {
	pathStr := "./track/transcript/data/sorted.gtf.gz"
	posStr := chrom + ":" + strconv.Itoa(start) + "-" + strconv.Itoa(end)
	genes, err := ReadGTF(pathStr, posStr)
	if err != nil {
		return nil, err
	}
	return genes, nil
}
