package common

import (
	"fmt"
	"io"
	"net/http"
)

func RequestBytes(url string, offset int, length int) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	rangeHeader := fmt.Sprintf("bytes=%d-%d", offset, offset+length-1)
	req.Header.Set("Range", rangeHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
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
