package bigbed

import (
	"errors"
	"gb-api/track/common"
)

func ReadBigBed(url string, chr string, start int, end int) ([]BigBedData, error) {
	bb := BigBed{}
	bb.URL = url

	// Load header
	err := bb.LoadHeader()
	if err != nil {
		return nil, errors.New("Failed to load BigWig header: " + err.Error())
	}

	// Load metadata
	metaData, err := common.RequestBytes(bb.URL, 64, int(bb.Header.FullDataOffset)-64+5)
	if err != nil {
		return nil, errors.New("Failed to request metadata: " + err.Error())
	}

	err = bb.LoadMetaData(metaData)
	if err != nil {
		return nil, errors.New("Failed to load metadata: " + err.Error())
	}

	data, err := bb.ReadBigBedData(chr, int32(start), int32(end))
	if err != nil {
		return nil, errors.New("Failed to read BigWig data: " + err.Error())
	}
	return data, nil
}
