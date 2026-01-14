package bigwig

import (
	"testing"
)

func TestResampleToWidth_Empty(t *testing.T) {
	data := []BigWigData{}
	result := ResampleToWidth(data, 100)
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d bins", len(result))
	}
}

func TestResampleToWidth_SingleBin(t *testing.T) {
	data := []BigWigData{
		{Chr: "chr1", Start: 0, End: 10, Value: 1.0},
		{Chr: "chr1", Start: 10, End: 20, Value: 5.0},
		{Chr: "chr1", Start: 20, End: 30, Value: 3.0},
	}

	result := ResampleToWidth(data, 1)

	if len(result) != 1 {
		t.Fatalf("Expected 1 bin, got %d", len(result))
	}

	// Should have max=5.0, min=1.0, x=0
	if result[0].X != 0 {
		t.Errorf("Expected x=0, got x=%d", result[0].X)
	}
	if result[0].Max != 5.0 {
		t.Errorf("Expected max=5.0, got max=%f", result[0].Max)
	}
	if result[0].Min != 1.0 {
		t.Errorf("Expected min=1.0, got min=%f", result[0].Min)
	}
}

func TestResampleToWidth_MultipleBins(t *testing.T) {
	// Range: 0-100, split into 2 bins
	// Bin 0 covers [0, 50), Bin 1 covers [50, 100]
	data := []BigWigData{
		{Chr: "chr1", Start: 0, End: 25, Value: 10.0},
		{Chr: "chr1", Start: 25, End: 50, Value: 20.0}, // End=50 touches bin 1 boundary
		{Chr: "chr1", Start: 50, End: 75, Value: 5.0},
		{Chr: "chr1", Start: 75, End: 100, Value: 15.0},
	}

	result := ResampleToWidth(data, 2)

	if len(result) != 2 {
		t.Fatalf("Expected 2 bins, got %d", len(result))
	}

	// Bin 0: 0-50, contains values 10.0 and 20.0
	if result[0].X != 0 {
		t.Errorf("Bin 0: Expected x=0, got x=%d", result[0].X)
	}
	if result[0].Max != 20.0 {
		t.Errorf("Bin 0: Expected max=20.0, got max=%f", result[0].Max)
	}
	if result[0].Min != 10.0 {
		t.Errorf("Bin 0: Expected min=10.0, got min=%f", result[0].Min)
	}

	// Bin 1: 50-100, contains values 5.0, 15.0, AND 20.0 (because point 25-50 touches boundary)
	// The point ending at 50 spans into bin 1, so max is 20.0
	if result[1].X != 1 {
		t.Errorf("Bin 1: Expected x=1, got x=%d", result[1].X)
	}
	if result[1].Max != 20.0 {
		t.Errorf("Bin 1: Expected max=20.0, got max=%f", result[1].Max)
	}
	if result[1].Min != 5.0 {
		t.Errorf("Bin 1: Expected min=5.0, got min=%f", result[1].Min)
	}
}

func TestResampleToWidth_RealWorld775To1000(t *testing.T) {
	// Simulate 775 data points with varying values
	data := make([]BigWigData, 775)
	for i := 0; i < 775; i++ {
		data[i] = BigWigData{
			Chr:   "chr19",
			Start: int32(i * 10),
			End:   int32((i + 1) * 10),
			Value: float32(i % 100), // values 0-99 repeating
		}
	}

	result := ResampleToWidth(data, 1000)

	if len(result) != 1000 {
		t.Fatalf("Expected 1000 bins, got %d", len(result))
	}

	// Verify first bin
	if result[0].X != 0 {
		t.Errorf("Expected first bin x=0, got x=%d", result[0].X)
	}

	// Verify last bin
	if result[999].X != 999 {
		t.Errorf("Expected last bin x=999, got x=%d", result[999].X)
	}

	// Each bin should have max >= min
	for i, bin := range result {
		if bin.Max < bin.Min {
			t.Errorf("Bin %d: max (%f) < min (%f)", i, bin.Max, bin.Min)
		}
	}
}

func TestResampleToWidth_MaxMinInSameBin(t *testing.T) {
	data := []BigWigData{
		{Chr: "chr1", Start: 0, End: 10, Value: 50.0},
		{Chr: "chr1", Start: 10, End: 20, Value: 100.0}, // max
		{Chr: "chr1", Start: 20, End: 30, Value: 25.0},  // min
		{Chr: "chr1", Start: 30, End: 40, Value: 75.0},
	}

	result := ResampleToWidth(data, 1)

	if len(result) != 1 {
		t.Fatalf("Expected 1 bin, got %d", len(result))
	}

	if result[0].Max != 100.0 {
		t.Errorf("Expected max=100.0, got max=%f", result[0].Max)
	}
	if result[0].Min != 25.0 {
		t.Errorf("Expected min=25.0, got min=%f", result[0].Min)
	}
}

func TestResampleToWidth_ExampleOutput(t *testing.T) {
	// Test the exact format from the user's example
	data := []BigWigData{
		{Chr: "chr1", Start: 0, End: 100, Value: 86.40404510498047},
		{Chr: "chr1", Start: 100, End: 200, Value: 85.44567108154297},
		{Chr: "chr1", Start: 200, End: 300, Value: 87.0},
	}

	result := ResampleToWidth(data, 3)

	if len(result) != 3 {
		t.Fatalf("Expected 3 bins, got %d", len(result))
	}

	// First bin should have x=0
	if result[0].X != 0 {
		t.Errorf("Expected x=0, got x=%d", result[0].X)
	}

	// Check that max and min are present
	if result[0].Max == 0 && result[0].Min == 0 {
		t.Error("Expected non-zero max and min values")
	}

	// Verify max >= min for all bins
	for i, bin := range result {
		if bin.Max < bin.Min {
			t.Errorf("Bin %d: max (%f) should be >= min (%f)", i, bin.Max, bin.Min)
		}
		if bin.X != i {
			t.Errorf("Bin %d: expected x=%d, got x=%d", i, i, bin.X)
		}
	}
}

func TestResampleToWidth_ZeroTarget(t *testing.T) {
	data := []BigWigData{
		{Chr: "chr1", Start: 0, End: 10, Value: 1.0},
	}

	result := ResampleToWidth(data, 0)

	// Should return empty when target is 0
	if len(result) != 0 {
		t.Errorf("Expected empty result for target=0, got %d bins", len(result))
	}
}
