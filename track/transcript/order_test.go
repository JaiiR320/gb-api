package transcript

import (
	"testing"
)

func TestPackTranscripts_Empty(t *testing.T) {
	result := PackTranscripts([]LegacyTranscript{}, 100)
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d items", len(result))
	}
}

func TestPackTranscripts_SingleTranscript(t *testing.T) {
	transcripts := []LegacyTranscript{
		{
			ID: "tx1",
			Coordinates: LegacyCoordinates{
				Start: 100,
				End:   200,
			},
		},
	}

	result := PackTranscripts(transcripts, 100)

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	if result["tx1"] != 0 {
		t.Errorf("Expected tx1 in row 0, got row %d", result["tx1"])
	}
}

func TestPackTranscripts_NonOverlapping(t *testing.T) {
	transcripts := []LegacyTranscript{
		{
			ID: "tx1",
			Coordinates: LegacyCoordinates{
				Start: 100,
				End:   200,
			},
		},
		{
			ID: "tx2",
			Coordinates: LegacyCoordinates{
				Start: 400,
				End:   500,
			},
		},
	}

	result := PackTranscripts(transcripts, 100)

	// Both should fit in row 0 since they don't overlap
	if result["tx1"] != 0 || result["tx2"] != 0 {
		t.Errorf("Expected both transcripts in row 0, got tx1=%d, tx2=%d", result["tx1"], result["tx2"])
	}
}

func TestPackTranscripts_Overlapping(t *testing.T) {
	transcripts := []LegacyTranscript{
		{
			ID: "tx1",
			Coordinates: LegacyCoordinates{
				Start: 100,
				End:   200,
			},
		},
		{
			ID: "tx2",
			Coordinates: LegacyCoordinates{
				Start: 150,
				End:   250,
			},
		},
	}

	result := PackTranscripts(transcripts, 100)

	// tx1 should be in row 0, tx2 in row 1 due to overlap
	if result["tx1"] != 0 {
		t.Errorf("Expected tx1 in row 0, got row %d", result["tx1"])
	}
	if result["tx2"] != 1 {
		t.Errorf("Expected tx2 in row 1, got row %d", result["tx2"])
	}
}

func TestPackTranscripts_WithPadding(t *testing.T) {
	transcripts := []LegacyTranscript{
		{
			ID: "tx1",
			Coordinates: LegacyCoordinates{
				Start: 100,
				End:   200,
			},
		},
		{
			ID: "tx2",
			Coordinates: LegacyCoordinates{
				Start: 250, // 50bp away, less than 2*padding
				End:   350,
			},
		},
	}

	// With paddingBp=100
	// tx1 layout: 0-300, tx2 layout: 150-450
	// They overlap, so should be in different rows
	result := PackTranscripts(transcripts, 100)

	if result["tx1"] != 0 {
		t.Errorf("Expected tx1 in row 0, got row %d", result["tx1"])
	}
	if result["tx2"] != 1 {
		t.Errorf("Expected tx2 in row 1 due to padding, got row %d", result["tx2"])
	}
}

func TestPackTranscripts_Deterministic(t *testing.T) {
	transcripts := []LegacyTranscript{
		{
			ID: "tx_b",
			Coordinates: LegacyCoordinates{
				Start: 100,
				End:   200,
			},
		},
		{
			ID: "tx_a",
			Coordinates: LegacyCoordinates{
				Start: 100,
				End:   200,
			},
		},
	}

	// Run multiple times to ensure determinism
	result1 := PackTranscripts(transcripts, 100)
	result2 := PackTranscripts(transcripts, 100)

	if result1["tx_a"] != result2["tx_a"] || result1["tx_b"] != result2["tx_b"] {
		t.Errorf("Results are not deterministic: run1=%v, run2=%v", result1, result2)
	}

	// tx_a should come first due to lexicographic sorting
	if result1["tx_a"] != 0 {
		t.Errorf("Expected tx_a in row 0, got row %d", result1["tx_a"])
	}
	if result1["tx_b"] != 1 {
		t.Errorf("Expected tx_b in row 1, got row %d", result1["tx_b"])
	}
}

func TestPackTranscripts_ComplexStacking(t *testing.T) {
	transcripts := []LegacyTranscript{
		{
			ID: "tx1",
			Coordinates: LegacyCoordinates{
				Start: 100,
				End:   200,
			},
		},
		{
			ID: "tx2",
			Coordinates: LegacyCoordinates{
				Start: 150,
				End:   250,
			},
		},
		{
			ID: "tx3",
			Coordinates: LegacyCoordinates{
				Start: 500,
				End:   600,
			},
		},
		{
			ID: "tx4",
			Coordinates: LegacyCoordinates{
				Start: 550,
				End:   650,
			},
		},
	}

	result := PackTranscripts(transcripts, 100)

	// tx1 and tx3 don't overlap, should be in row 0
	// tx2 and tx4 don't overlap, should be in row 1
	if result["tx1"] != 0 {
		t.Errorf("Expected tx1 in row 0, got row %d", result["tx1"])
	}
	if result["tx2"] != 1 {
		t.Errorf("Expected tx2 in row 1, got row %d", result["tx2"])
	}
	if result["tx3"] != 0 {
		t.Errorf("Expected tx3 in row 0, got row %d", result["tx3"])
	}
	if result["tx4"] != 1 {
		t.Errorf("Expected tx4 in row 1, got row %d", result["tx4"])
	}
}

func TestPackTranscripts_DifferentPadding(t *testing.T) {
	transcripts := []LegacyTranscript{
		{
			ID: "tx1",
			Coordinates: LegacyCoordinates{
				Start: 100,
				End:   200,
			},
		},
		{
			ID: "tx2",
			Coordinates: LegacyCoordinates{
				Start: 250,
				End:   350,
			},
		},
	}

	// With small padding, they fit together
	resultSmallPadding := PackTranscripts(transcripts, 10)
	// tx1: 90-210, tx2: 240-360 -> fit in same row

	// With large padding, they need separate rows
	resultLargePadding := PackTranscripts(transcripts, 100)
	// tx1: 0-300, tx2: 150-450 -> overlap, different rows

	if resultSmallPadding["tx1"] != 0 || resultSmallPadding["tx2"] != 0 {
		t.Errorf("Expected both in row 0 with small padding, got tx1=%d, tx2=%d",
			resultSmallPadding["tx1"], resultSmallPadding["tx2"])
	}

	if resultLargePadding["tx1"] != 0 || resultLargePadding["tx2"] != 1 {
		t.Errorf("Expected different rows with large padding, got tx1=%d, tx2=%d",
			resultLargePadding["tx1"], resultLargePadding["tx2"])
	}
}

func TestPackTranscriptsWithLayout(t *testing.T) {
	transcripts := []LegacyTranscript{
		{
			ID: "tx1",
			Coordinates: LegacyCoordinates{
				Start: 100,
				End:   200,
			},
		},
		{
			ID: "tx2",
			Coordinates: LegacyCoordinates{
				Start: 150,
				End:   250,
			},
		},
	}

	result := PackTranscriptsWithLayout(transcripts, 100)

	if len(result) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result))
	}

	// Check that IDs are preserved
	if result[0].ID != "tx1" || result[1].ID != "tx2" {
		t.Errorf("Expected IDs to be preserved in order")
	}

	// Check row assignments
	if result[0].Row != 0 || result[1].Row != 1 {
		t.Errorf("Expected tx1 in row 0 and tx2 in row 1, got %d and %d",
			result[0].Row, result[1].Row)
	}
}

func TestGetTotalRows(t *testing.T) {
	tests := []struct {
		name     string
		layout   map[string]int
		expected int
	}{
		{
			name:     "empty layout",
			layout:   map[string]int{},
			expected: 0,
		},
		{
			name: "single row",
			layout: map[string]int{
				"tx1": 0,
				"tx2": 0,
			},
			expected: 1,
		},
		{
			name: "multiple rows",
			layout: map[string]int{
				"tx1": 0,
				"tx2": 1,
				"tx3": 2,
			},
			expected: 3,
		},
		{
			name: "non-contiguous rows",
			layout: map[string]int{
				"tx1": 0,
				"tx2": 5,
			},
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTotalRows(tt.layout)
			if result != tt.expected {
				t.Errorf("Expected %d rows, got %d", tt.expected, result)
			}
		})
	}
}
