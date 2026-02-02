package bigdata

import (
	"bytes"
	"compress/zlib"
	"sync"
	"testing"
)

// compressData is a test helper to create zlib-compressed data
func compressData(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		t.Fatalf("failed to compress test data: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zlib writer: %v", err)
	}
	return buf.Bytes()
}

func TestDecompressData_NoDecompression(t *testing.T) {
	input := []byte("hello world")
	result, err := DecompressData(input, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !bytes.Equal(result, input) {
		t.Errorf("got %v, want %v", result, input)
	}
}

func TestDecompressData_BasicDecompression(t *testing.T) {
	original := []byte("hello world, this is a test of decompression")
	compressed := compressData(t, original)

	result, err := DecompressData(compressed, true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !bytes.Equal(result, original) {
		t.Errorf("got %q, want %q", result, original)
	}
}

func TestDecompressData_LargeData(t *testing.T) {
	// Create data larger than the default buffer capacity (64KB)
	original := make([]byte, 128*1024) // 128KB
	for i := range original {
		original[i] = byte(i % 256)
	}
	compressed := compressData(t, original)

	result, err := DecompressData(compressed, true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !bytes.Equal(result, original) {
		t.Errorf("decompressed data mismatch for large data")
	}
}

func TestDecompressData_InvalidData(t *testing.T) {
	invalid := []byte("not valid zlib data")
	_, err := DecompressData(invalid, true)
	if err == nil {
		t.Error("expected error for invalid compressed data, got nil")
	}
}

func TestDecompressData_EmptyData(t *testing.T) {
	original := []byte{}
	compressed := compressData(t, original)

	result, err := DecompressData(compressed, true)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d bytes", len(result))
	}
}

func TestDecompressData_Concurrent(t *testing.T) {
	// Test that buffer pool works correctly under concurrent access
	original := []byte("test data for concurrent decompression testing")
	compressed := compressData(t, original)

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			result, err := DecompressData(compressed, true)
			if err != nil {
				errors <- err
				return
			}
			if !bytes.Equal(result, original) {
				errors <- err
				return
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent decompression error: %v", err)
	}
}

func TestDecompressData_BufferReuse(t *testing.T) {
	// Verify that multiple sequential decompressions work correctly
	// (tests that buffer Reset() works properly)
	tests := []struct {
		name string
		data string
	}{
		{name: "first", data: "first decompression"},
		{name: "second", data: "second decompression with different data"},
		{name: "third", data: "third time is the charm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := []byte(tt.data)
			compressed := compressData(t, original)

			result, err := DecompressData(compressed, true)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !bytes.Equal(result, original) {
				t.Errorf("got %q, want %q", result, original)
			}
		})
	}
}
