package transcript

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

var projectRoot string

func init() {
	// Find project root by looking for go.mod, starting from this file's directory
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			projectRoot = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			break
		}
		dir = parent
	}

	// Change to project root so GetTranscripts can find the data file
	// GetTranscripts uses path "./track/transcript/data/v40/sorted.gtf.gz"
	if projectRoot != "" {
		os.Chdir(projectRoot)
	}
}

// skipIfNoGTFData skips the test if the GTF data file is not accessible.
func skipIfNoGTFData(t *testing.T) {
	t.Helper()
	gtfPath := "./track/transcript/data/v40/sorted.gtf.gz"
	if _, err := os.Stat(gtfPath); os.IsNotExist(err) {
		t.Skipf("GTF data file not found at %s (cwd: %s)", gtfPath, mustGetwd())
	}
}

func mustGetwd() string {
	wd, _ := os.Getwd()
	return wd
}

func TestGetTranscripts(t *testing.T) {
	tests := []struct {
		name     string
		chrom    string
		start    int
		end      int
		wantErr  bool
		validate func(t *testing.T, genes []Gene)
	}{
		{
			name:    "APOE region - known gene",
			chrom:   "chr19",
			start:   44905754,
			end:     44907754,
			wantErr: false,
			validate: func(t *testing.T, genes []Gene) {
				if len(genes) == 0 {
					t.Error("expected genes in APOE region, got none")
				}
				// Look for APOE gene
				foundAPOE := false
				for _, g := range genes {
					if g.Name == "APOE" {
						foundAPOE = true
						if g.Strand != "+" && g.Strand != "-" {
							t.Errorf("invalid strand for APOE: %s", g.Strand)
						}
						if len(g.Transcripts) == 0 {
							t.Error("APOE should have transcripts")
						}
					}
				}
				if !foundAPOE {
					t.Logf("genes found: %v", geneNames(genes))
				}
			},
		},
		{
			name:    "larger region with multiple genes",
			chrom:   "chr19",
			start:   44900000,
			end:     44920000,
			wantErr: false,
			validate: func(t *testing.T, genes []Gene) {
				if len(genes) == 0 {
					t.Error("expected genes in region, got none")
				}
				// Verify gene structure
				for _, g := range genes {
					if g.ID == "" {
						t.Errorf("gene %s has empty ID", g.Name)
					}
					if g.Chrom != "chr19" {
						t.Errorf("expected chr19, got %s", g.Chrom)
					}
				}
			},
		},
		{
			name:    "region with transcripts and exons",
			chrom:   "chr19",
			start:   44905000,
			end:     44910000,
			wantErr: false,
			validate: func(t *testing.T, genes []Gene) {
				for _, g := range genes {
					for _, tr := range g.Transcripts {
						if tr.ID == "" {
							t.Errorf("transcript %s has empty ID", tr.Name)
						}
						if len(tr.Exons) == 0 {
							t.Errorf("transcript %s has no exons", tr.Name)
						}
						// Verify exon structure
						for _, ex := range tr.Exons {
							if ex.ExonNumber < 1 {
								t.Errorf("invalid exon number: %d", ex.ExonNumber)
							}
							if ex.Start >= ex.End {
								t.Errorf("invalid exon range: %d-%d", ex.Start, ex.End)
							}
						}
					}
				}
			},
		},
		{
			name:    "empty region",
			chrom:   "chr1",
			start:   1,
			end:     100,
			wantErr: false,
			validate: func(t *testing.T, genes []Gene) {
				// Empty result is valid for regions without annotations
			},
		},
		{
			name:    "chr1 known region",
			chrom:   "chr1",
			start:   11873,
			end:     14409,
			wantErr: false,
			validate: func(t *testing.T, genes []Gene) {
				// This region should have DDX11L1 or similar
				// Just verify structure if genes exist
				for _, g := range genes {
					if g.Start >= g.End {
						t.Errorf("invalid gene range: %d-%d", g.Start, g.End)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipIfNoGTFData(t)
			genes, err := GetTranscripts(tt.chrom, tt.start, tt.end)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTranscripts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, genes)
			}
		})
	}
}

func TestGetTranscriptsGeneStructure(t *testing.T) {
	skipIfNoGTFData(t)
	genes, err := GetTranscripts("chr19", 44905000, 44910000)
	if err != nil {
		t.Fatalf("GetTranscripts() error = %v", err)
	}

	if len(genes) == 0 {
		t.Skip("no genes in test region")
	}

	for _, gene := range genes {
		// Test Gene fields
		if gene.Name == "" {
			t.Error("gene name should not be empty")
		}
		if gene.ID == "" {
			t.Error("gene ID should not be empty")
		}
		if gene.Type == "" {
			t.Error("gene type should not be empty")
		}
		if gene.Strand != "+" && gene.Strand != "-" {
			t.Errorf("invalid strand: %s", gene.Strand)
		}

		// Test Transcript fields
		for _, tr := range gene.Transcripts {
			if tr.Name == "" {
				t.Error("transcript name should not be empty")
			}
			if tr.ID == "" {
				t.Error("transcript ID should not be empty")
			}

			// Test Exon fields
			for _, ex := range tr.Exons {
				if ex.ExonNumber < 1 {
					t.Errorf("exon number should be >= 1, got %d", ex.ExonNumber)
				}
				if ex.Chrom != gene.Chrom {
					t.Errorf("exon chrom %s doesn't match gene chrom %s", ex.Chrom, gene.Chrom)
				}
			}
		}
	}
}

func TestGetTranscriptsCanonicalTranscript(t *testing.T) {
	skipIfNoGTFData(t)
	// Test that canonical transcripts are properly identified
	genes, err := GetTranscripts("chr19", 44905000, 44910000)
	if err != nil {
		t.Fatalf("GetTranscripts() error = %v", err)
	}

	canonicalCount := 0
	for _, gene := range genes {
		for _, tr := range gene.Transcripts {
			if tr.Canonical {
				canonicalCount++
				t.Logf("Found canonical transcript: %s for gene %s", tr.Name, gene.Name)
			}
		}
	}

	// Log for information - not all genes have MANE_Select transcripts
	t.Logf("Found %d canonical transcripts in %d genes", canonicalCount, len(genes))
}

func TestGetTranscriptsCoordinates(t *testing.T) {
	skipIfNoGTFData(t)
	// Verify that returned genes overlap with the requested region
	start := 44905000
	end := 44910000

	genes, err := GetTranscripts("chr19", start, end)
	if err != nil {
		t.Fatalf("GetTranscripts() error = %v", err)
	}

	for _, gene := range genes {
		// Gene should overlap with requested region
		if gene.End < start || gene.Start > end {
			t.Errorf("gene %s (%d-%d) doesn't overlap requested region (%d-%d)",
				gene.Name, gene.Start, gene.End, start, end)
		}
	}
}

// Helper function to extract gene names for logging
func geneNames(genes []Gene) []string {
	names := make([]string, len(genes))
	for i, g := range genes {
		names[i] = g.Name
	}
	return names
}
