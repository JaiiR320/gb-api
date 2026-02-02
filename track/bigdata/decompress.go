package bigdata

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"sync"
)

const DEFAULT_BUFFER_CAPACITY = 64 * 1024 // 64KB default capacity

// bufferPool reuses decompression buffers to reduce GC pressure
var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, DEFAULT_BUFFER_CAPACITY))
	},
}

// DecompressData decompresses zlib-compressed data if needed
func DecompressData(data []byte, needsDecompression bool) ([]byte, error) {
	if !needsDecompression {
		return data, nil
	}

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer reader.Close()

	// Get buffer from pool and reset it
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Copy decompressed data into pooled buffer
	if _, err := io.Copy(buf, reader); err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// Copy result out before returning (buffer will be reused)
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	return result, nil
}
