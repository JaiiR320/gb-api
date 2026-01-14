package bigwig

import (
	"testing"
)

const testBigWigURL = "https://downloads.wenglab.org/DNAse_All_ENCODE_MAR20_2024_merged.bw"

func TestReadBigWig(t *testing.T) {
	tests := []struct {
		name             string
		url              string
		chr              string
		start            int
		end              int
		preRenderedWidth int
		wantErr          bool
		validate         func(t *testing.T, data []BigWigData)
	}{
		{
			name:             "valid region with data",
			url:              testBigWigURL,
			chr:              "chr19",
			start:            44905000,
			end:              44916000,
			preRenderedWidth: 0,
			wantErr:          false,
			validate: func(t *testing.T, data []BigWigData) {
				if len(data) == 0 {
					t.Error("expected data points, got none")
				}
				// Verify all data points are within requested range
				for _, d := range data {
					if d.Chr != "chr19" {
						t.Errorf("expected chr19, got %s", d.Chr)
					}
					if d.End < 44905000 || d.Start > 44916000 {
						t.Errorf("data point outside requested range: %d-%d", d.Start, d.End)
					}
				}
			},
		},
		{
			name:             "valid region with prerendered width",
			url:              testBigWigURL,
			chr:              "chr19",
			start:            44905000,
			end:              44916000,
			preRenderedWidth: 1000,
			wantErr:          false,
			validate: func(t *testing.T, data []BigWigData) {
				if len(data) == 0 {
					t.Error("expected data points, got none")
				}
			},
		},
		{
			name:             "small region",
			url:              testBigWigURL,
			chr:              "chr19",
			start:            44905000,
			end:              44905500,
			preRenderedWidth: 0,
			wantErr:          false,
			validate: func(t *testing.T, data []BigWigData) {
				// Small regions may have no data, that's okay
				for _, d := range data {
					if d.Chr != "chr19" {
						t.Errorf("expected chr19, got %s", d.Chr)
					}
				}
			},
		},
		{
			name:             "invalid chromosome",
			url:              testBigWigURL,
			chr:              "chrINVALID",
			start:            1000,
			end:              2000,
			preRenderedWidth: 0,
			wantErr:          true,
			validate:         nil,
		},
		{
			name:             "invalid URL",
			url:              "https://invalid.example.com/nonexistent.bw",
			chr:              "chr1",
			start:            1000,
			end:              2000,
			preRenderedWidth: 0,
			wantErr:          true,
			validate:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ReadBigWig(tt.url, tt.chr, tt.start, tt.end, tt.preRenderedWidth)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadBigWig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, data)
			}
		})
	}
}

func TestReadBigWigDataValues(t *testing.T) {
	// Test that we get expected values for a known region
	data, err := ReadBigWig(testBigWigURL, "chr19", 44905740, 44905800, 0)
	if err != nil {
		t.Fatalf("ReadBigWig() error = %v", err)
	}

	if len(data) == 0 {
		t.Skip("no data in test region, skipping value validation")
	}

	// Verify data structure
	for _, d := range data {
		if d.Start >= d.End {
			t.Errorf("invalid range: start %d >= end %d", d.Start, d.End)
		}
		// Values should be non-negative for this signal track
		if d.Value < 0 {
			t.Errorf("unexpected negative value: %f", d.Value)
		}
	}
}

func TestReadBigWigEmptyRegion(t *testing.T) {
	// Test a region that likely has no signal
	data, err := ReadBigWig(testBigWigURL, "chr1", 1, 100, 0)
	if err != nil {
		t.Fatalf("ReadBigWig() error = %v", err)
	}

	// Empty result is valid - just verify no error occurred
	_ = data
}
