package bigwig

import (
	"bytes"
	"encoding/binary"
	"gb-api/track/bigdata"
	"testing"
)

func TestDecodeZoomData_SingleRecord(t *testing.T) {
	// Create mock BigData with chromosome mapping
	b := &bigdata.BigData{
		ByteOrder: binary.LittleEndian,
		ChromTree: bigdata.ChromTree{
			IDToChrom: map[int32]string{0: "chr1"},
		},
	}

	// Create zoom record: chr1:100-200, mean value = 50
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int32(0))       // chromId
	binary.Write(buf, binary.LittleEndian, int32(100))     // start
	binary.Write(buf, binary.LittleEndian, int32(200))     // end
	binary.Write(buf, binary.LittleEndian, uint32(10))     // validCount
	binary.Write(buf, binary.LittleEndian, float32(40.0))  // minVal
	binary.Write(buf, binary.LittleEndian, float32(60.0))  // maxVal
	binary.Write(buf, binary.LittleEndian, float32(500.0)) // sumData
	binary.Write(buf, binary.LittleEndian, float32(0))     // sumSquares

	data, err := decodeZoomData(b, buf.Bytes(), 0, 0, 0, 1000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(data) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(data))
	}

	// Mean = sumData / validCount = 500 / 10 = 50
	expectedValue := float32(50.0)
	if data[0].Value != expectedValue {
		t.Errorf("Expected value %f, got %f", expectedValue, data[0].Value)
	}

	if data[0].Start != 100 || data[0].End != 200 {
		t.Errorf("Expected range 100-200, got %d-%d", data[0].Start, data[0].End)
	}

	if data[0].Chr != "chr1" {
		t.Errorf("Expected chr1, got %s", data[0].Chr)
	}
}

func TestDecodeZoomData_ZeroValidCount(t *testing.T) {
	b := &bigdata.BigData{
		ByteOrder: binary.LittleEndian,
		ChromTree: bigdata.ChromTree{
			IDToChrom: map[int32]string{0: "chr1"},
		},
	}

	// Create zoom record with validCount=0
	buf := new(bytes.Buffer)
	writeZoomRecord(buf, 0, 100, 200, 0, 500.0)

	data, err := decodeZoomData(b, buf.Bytes(), 0, 0, 0, 1000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(data) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(data))
	}

	// Value should be 0 when validCount is 0
	if data[0].Value != 0 {
		t.Errorf("Expected value 0 for validCount=0, got %f", data[0].Value)
	}
}

func TestDecodeZoomData_Filtering(t *testing.T) {
	b := &bigdata.BigData{
		ByteOrder: binary.LittleEndian,
		ChromTree: bigdata.ChromTree{
			IDToChrom: map[int32]string{0: "chr1"},
		},
	}

	// Create 3 zoom records
	buf := new(bytes.Buffer)

	// Record 1: 0-100 (before filter)
	writeZoomRecord(buf, 0, 0, 100, 10, 100.0)

	// Record 2: 500-600 (within filter)
	writeZoomRecord(buf, 0, 500, 600, 10, 200.0)

	// Record 3: 1500-1600 (after filter)
	writeZoomRecord(buf, 0, 1500, 1600, 10, 300.0)

	// Filter: chr1:400-1000
	data, err := decodeZoomData(b, buf.Bytes(), 0, 400, 0, 1000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Only record 2 should be included
	if len(data) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(data))
	}

	if data[0].Start != 500 {
		t.Errorf("Expected start 500, got %d", data[0].Start)
	}

	// Mean = 200 / 10 = 20
	expectedValue := float32(20.0)
	if data[0].Value != expectedValue {
		t.Errorf("Expected value %f, got %f", expectedValue, data[0].Value)
	}
}

func TestDecodeZoomData_MultipleRecords(t *testing.T) {
	b := &bigdata.BigData{
		ByteOrder: binary.LittleEndian,
		ChromTree: bigdata.ChromTree{
			IDToChrom: map[int32]string{0: "chr1"},
		},
	}

	// Create 3 zoom records all within the filter range
	buf := new(bytes.Buffer)
	writeZoomRecord(buf, 0, 100, 200, 10, 100.0)  // mean = 10
	writeZoomRecord(buf, 0, 200, 300, 20, 600.0)  // mean = 30
	writeZoomRecord(buf, 0, 300, 400, 5, 250.0)   // mean = 50

	// Filter: chr1:0-500 (includes all)
	data, err := decodeZoomData(b, buf.Bytes(), 0, 0, 0, 500)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(data) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(data))
	}

	// Verify each record
	expectedValues := []float32{10.0, 30.0, 50.0}
	for i, expected := range expectedValues {
		if data[i].Value != expected {
			t.Errorf("Record %d: expected value %f, got %f", i, expected, data[i].Value)
		}
	}
}

func TestDecodeZoomData_ChromosomeFiltering(t *testing.T) {
	b := &bigdata.BigData{
		ByteOrder: binary.LittleEndian,
		ChromTree: bigdata.ChromTree{
			IDToChrom: map[int32]string{
				0: "chr1",
				1: "chr2",
			},
		},
	}

	// Create records for different chromosomes
	buf := new(bytes.Buffer)
	writeZoomRecord(buf, 0, 100, 200, 10, 100.0) // chr1
	writeZoomRecord(buf, 1, 100, 200, 10, 200.0) // chr2
	writeZoomRecord(buf, 0, 300, 400, 10, 300.0) // chr1

	// Filter: only chr1 (chromId=0)
	data, err := decodeZoomData(b, buf.Bytes(), 0, 0, 0, 500)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should get 2 chr1 records
	if len(data) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(data))
	}

	// Verify both are chr1
	for i, record := range data {
		if record.Chr != "chr1" {
			t.Errorf("Record %d: expected chr1, got %s", i, record.Chr)
		}
	}
}

// Helper function to write a zoom record to a buffer
func writeZoomRecord(buf *bytes.Buffer, chromId, start, end int32, validCount uint32, sumData float32) {
	binary.Write(buf, binary.LittleEndian, chromId)
	binary.Write(buf, binary.LittleEndian, start)
	binary.Write(buf, binary.LittleEndian, end)
	binary.Write(buf, binary.LittleEndian, validCount)
	binary.Write(buf, binary.LittleEndian, float32(0))   // minVal
	binary.Write(buf, binary.LittleEndian, float32(0))   // maxVal
	binary.Write(buf, binary.LittleEndian, sumData)
	binary.Write(buf, binary.LittleEndian, float32(0))   // sumSquares
}
