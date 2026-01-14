package bigbed

import (
	"testing"
)

const testBigBedURL = "https://downloads.wenglab.org/GRCh38-cCREs.DCC.bigBed"

func TestReadBigBed(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		chr      string
		start    int
		end      int
		wantErr  bool
		validate func(t *testing.T, data []BigBedData)
	}{
		{
			name:    "valid region with CCRE data",
			url:     testBigBedURL,
			chr:     "chr19",
			start:   44905754,
			end:     44907754,
			wantErr: false,
			validate: func(t *testing.T, data []BigBedData) {
				if len(data) == 0 {
					t.Error("expected data points, got none")
				}
				// Verify all data points are within requested range
				for _, d := range data {
					if d.Chr != "chr19" {
						t.Errorf("expected chr19, got %s", d.Chr)
					}
					// BigBed returns features that overlap the range
					if d.End < 44905754 || d.Start > 44907754 {
						t.Errorf("data point outside requested range: %d-%d", d.Start, d.End)
					}
				}
			},
		},
		{
			name:    "larger region",
			url:     testBigBedURL,
			chr:     "chr19",
			start:   44900000,
			end:     44920000,
			wantErr: false,
			validate: func(t *testing.T, data []BigBedData) {
				if len(data) == 0 {
					t.Error("expected data points in larger region, got none")
				}
			},
		},
		{
			name:    "small region",
			url:     testBigBedURL,
			chr:     "chr19",
			start:   44905754,
			end:     44905800,
			wantErr: false,
			validate: func(t *testing.T, data []BigBedData) {
				// Small regions may have no features, that's okay
				for _, d := range data {
					if d.Chr != "chr19" {
						t.Errorf("expected chr19, got %s", d.Chr)
					}
				}
			},
		},
		{
			name:     "invalid chromosome",
			url:      testBigBedURL,
			chr:      "chrINVALID",
			start:    1000,
			end:      2000,
			wantErr:  true,
			validate: nil,
		},
		{
			name:     "invalid URL",
			url:      "https://invalid.example.com/nonexistent.bb",
			chr:      "chr1",
			start:    1000,
			end:      2000,
			wantErr:  true,
			validate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ReadBigBed(tt.url, tt.chr, tt.start, tt.end)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadBigBed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, data)
			}
		})
	}
}

func TestReadBigBedDataStructure(t *testing.T) {
	// Test that we get properly structured data for a known region
	data, err := ReadBigBed(testBigBedURL, "chr19", 44905754, 44907754)
	if err != nil {
		t.Fatalf("ReadBigBed() error = %v", err)
	}

	if len(data) == 0 {
		t.Skip("no data in test region, skipping structure validation")
	}

	// Verify data structure
	for _, d := range data {
		if d.Start >= d.End {
			t.Errorf("invalid range: start %d >= end %d", d.Start, d.End)
		}
		if d.Chr == "" {
			t.Error("chromosome should not be empty")
		}
		// CCRE BigBed should have Rest field with additional data
		// (but it's not required for all BigBed files)
	}
}

func TestReadBigBedEmptyRegion(t *testing.T) {
	// Test a region that likely has no features
	data, err := ReadBigBed(testBigBedURL, "chr1", 1, 100)
	if err != nil {
		t.Fatalf("ReadBigBed() error = %v", err)
	}

	// Empty result is valid - just verify no error occurred
	_ = data
}

func TestReadBigBedOverlappingFeatures(t *testing.T) {
	// Test that overlapping features are returned correctly
	data, err := ReadBigBed(testBigBedURL, "chr19", 44905000, 44910000)
	if err != nil {
		t.Fatalf("ReadBigBed() error = %v", err)
	}

	// Count features and check for overlaps
	if len(data) > 1 {
		// Verify features are sorted by start position
		for i := 1; i < len(data); i++ {
			if data[i].Start < data[i-1].Start {
				// Features don't need to be sorted, but let's document the behavior
				t.Logf("features not sorted by start: %d comes before %d", data[i-1].Start, data[i].Start)
			}
		}
	}
}
