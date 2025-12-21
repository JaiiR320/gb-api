package bigdata

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
)

// decompressData decompresses zlib-compressed data if needed
func DecompressData(data []byte, needsDecompression bool) ([]byte, error) {
	if !needsDecompression {
		return data, nil
	}

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return decompressed, nil
}
