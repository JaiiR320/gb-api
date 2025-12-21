package bigwig

import (
	"errors"
	"gb-api/track/common"
)

func ReadBigWig(url string, chrom string, start int, end int) ([]BigWigData, error) {
	bw := BigWig{}
	bw.URL = url

	// Load header
	err := bw.LoadHeader()
	if err != nil {
		return nil, errors.New("Failed to load BigWig header: " + err.Error())
	}

	// Load metadata
	metaData, err := common.RequestBytes(bw.URL, 64, int(bw.Header.FullDataOffset)-64+5)
	if err != nil {
		return nil, errors.New("Failed to request metadata: " + err.Error())
	}

	err = bw.LoadMetaData(metaData)
	if err != nil {
		return nil, errors.New("Failed to load metadata: " + err.Error())
	}

	// Read BigWig data
	data, err := bw.ReadBigWigData(chrom, int32(start), chrom, int32(end))
	if err != nil {
		return nil, errors.New("Failed to read BigWig data: " + err.Error())
	}
	return data, nil
}
