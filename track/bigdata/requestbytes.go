package bigdata

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true, // Handle compression manually
	},
}

func RequestBytes(url string, offset int, length int) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	rangeHeader := fmt.Sprintf("bytes=%d-%d", offset, offset+length-1)
	req.Header.Set("Range", rangeHeader)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := make([]byte, length)
	_, err = io.ReadFull(resp.Body, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
