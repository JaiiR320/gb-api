package transcript

import (
	"strconv"
)

func GetTranscripts(chrom string, start int, end int) ([]Gene, error) {
	pathStr := "./track/transcript/data/v40/sorted.gtf.gz"
	posStr := chrom + ":" + strconv.Itoa(start) + "-" + strconv.Itoa(end)
	genes, err := ReadGTF(pathStr, posStr)
	if err != nil {
		return nil, err
	}
	return genes, nil
}

// Legacy types for backwards compatibility
type LegacyData struct {
	Gene []LegacyGene `json:"gene"`
}

type LegacyDataWithLayout struct {
	Gene      []LegacyGene `json:"gene"`
	TotalRows int          `json:"totalRows,omitempty"`
	PaddingBp int          `json:"paddingBp,omitempty"`
}

type LegacyGene struct {
	Strand      string             `json:"strand"`
	Name        string             `json:"name"`
	ID          string             `json:"id"`
	Transcripts []LegacyTranscript `json:"transcripts"`
	Typename    string             `json:"__typename"`
}

type LegacyTranscript struct {
	Coordinates LegacyCoordinates `json:"coordinates"`
	Name        string            `json:"name"`
	ID          string            `json:"id"`
	Exons       []LegacyExon      `json:"exons"`
	Typename    string            `json:"__typename"`
	Row         *int              `json:"row,omitempty"` // Row index for stacking (0 = top row)
}

type LegacyExon struct {
	Coordinates LegacyCoordinates `json:"coordinates"`
	UTRs        []LegacyUTR       `json:"UTRs"`
	Typename    string            `json:"__typename"`
}

type LegacyUTR struct {
	Coordinates LegacyCoordinates `json:"coordinates"`
	Typename    string            `json:"__typename"`
}

type LegacyCoordinates struct {
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Typename string `json:"__typename"`
}

// Legacy transforms genes into the legacy response format
func Legacy(genes []Gene, err error) (any, error) {
	return LegacyWithLayout(genes, 0, err)
}

// LegacyWithLayout transforms genes into the legacy response format with row ordering.
// If paddingBp is 0, a default of 100bp is used.
func LegacyWithLayout(genes []Gene, paddingBp int, err error) (any, error) {
	if err != nil {
		return nil, err
	}

	legacyGenes := make([]LegacyGene, 0, len(genes))
	allTranscripts := make([]LegacyTranscript, 0)

	// Build legacy structure
	for _, gene := range genes {
		legacyGene := LegacyGene{
			Strand:   gene.Strand,
			Name:     gene.Name,
			ID:       gene.ID,
			Typename: "Gene",
		}

		legacyTranscripts := make([]LegacyTranscript, 0, len(gene.Transcripts))
		for _, transcript := range gene.Transcripts {
			legacyTranscript := LegacyTranscript{
				Coordinates: LegacyCoordinates{
					Start:    transcript.Start,
					End:      transcript.End,
					Typename: "GenomicRange",
				},
				Name:     transcript.Name,
				ID:       transcript.ID,
				Typename: "Transcript",
			}

			legacyExons := make([]LegacyExon, 0, len(transcript.Exons))
			for _, exon := range transcript.Exons {
				legacyExon := LegacyExon{
					Coordinates: LegacyCoordinates{
						Start:    exon.Start,
						End:      exon.End,
						Typename: "GenomicRange",
					},
					UTRs:     make([]LegacyUTR, 0, len(exon.UTRs)),
					Typename: "Exon",
				}

				for _, utr := range exon.UTRs {
					legacyExon.UTRs = append(legacyExon.UTRs, LegacyUTR{
						Coordinates: LegacyCoordinates{
							Start:    utr.Start,
							End:      utr.End,
							Typename: "GenomicRange",
						},
						Typename: "UTR",
					})
				}

				legacyExons = append(legacyExons, legacyExon)
			}
			legacyTranscript.Exons = legacyExons
			legacyTranscripts = append(legacyTranscripts, legacyTranscript)
			allTranscripts = append(allTranscripts, legacyTranscript)
		}
		legacyGene.Transcripts = legacyTranscripts
		legacyGenes = append(legacyGenes, legacyGene)
	}

	if paddingBp <= 0 {
		paddingBp = 100 // Default padding in base pairs
	}

	// Compute row layout for all transcripts
	rowMap := PackTranscripts(allTranscripts, paddingBp)

	// Assign rows back to transcripts in genes
	for i := range legacyGenes {
		for j := range legacyGenes[i].Transcripts {
			txID := legacyGenes[i].Transcripts[j].ID
			if row, exists := rowMap[txID]; exists {
				legacyGenes[i].Transcripts[j].Row = &row
			}
		}
	}

	return LegacyDataWithLayout{
		Gene:      legacyGenes,
		TotalRows: GetTotalRows(rowMap),
		PaddingBp: paddingBp,
	}, nil
}
