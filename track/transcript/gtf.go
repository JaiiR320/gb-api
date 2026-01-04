package transcript

import (
	"fmt"
	"sort"
	"strconv"
)

type GenomicRange struct {
	Chrom string `json:"chrom"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type Feature struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	GenomicRange
}

type Gene struct {
	Feature
	Strand      string       `json:"strand"`
	Type        string       `json:"type"`
	Transcripts []Transcript `json:"transcripts"`
}

type Transcript struct {
	Feature
	Exons []Exon `json:"exons"`
}

type Exon struct {
	Feature
	ExonNumber int            `json:"exon_number"`
	UTRs       []GenomicRange `json:"utrs,omitempty"`
	CDSs       []GenomicRange `json:"cdss,omitempty"`
	StartCodon *GenomicRange  `json:"start_codon,omitempty"`
	StopCodon  *GenomicRange  `json:"stop_codon,omitempty"`
}

func ReadGTF(filePath string, posStr string) ([]Gene, error) {
	records, err := GetRecords(filePath, posStr)
	if err != nil {
		return nil, err
	}
	return buildGenes(records)
}

func buildGenes(records []Record) ([]Gene, error) {
	var Genes []Gene
	genesByName := filterByAttribute(records, "gene_name")

	// Sort gene names for deterministic order
	geneNames := make([]string, 0, len(genesByName))
	for geneName := range genesByName {
		geneNames = append(geneNames, geneName)
	}
	sort.Strings(geneNames)

	for _, geneName := range geneNames {
		geneObj, err := buildGene(geneName, genesByName[geneName])
		if err != nil {
			return nil, err
		}
		Genes = append(Genes, geneObj)
	}
	return Genes, nil
}

func buildExons(transcriptRecords []Record, startCodon, stopCodon []Record, strand string) ([]Exon, error) {
	var exons []Exon

	// Get all exon feature records
	exonRecords := filterByFeature(transcriptRecords, "exon")

	// Pre-filter UTRs and CDSs once (optimization)
	allUTRs := filterByFeature(transcriptRecords, "UTR")
	utrsByExonNumber := filterByAttribute(allUTRs, "exon_number")

	allCDSs := filterByFeature(transcriptRecords, "CDS")
	cdsByExonNumber := filterByAttribute(allCDSs, "exon_number")

	for _, exonRecord := range exonRecords {
		// Parse exon number
		exonNumber := exonRecord.Attributes["exon_number"]
		exonNumberValue, err := strconv.Atoi(exonNumber)
		if err != nil {
			return nil, fmt.Errorf("invalid exon number: %v", err)
		}

		// Create Exon object
		exonObj := Exon{
			ExonNumber: exonNumberValue,
			Feature: Feature{
				ID: exonRecord.Attributes["exon_id"],
				GenomicRange: GenomicRange{
					Chrom: exonRecord.Chrom,
					Start: exonRecord.Start,
					End:   exonRecord.End,
				},
			},
		}

		// Assign codons based on strand direction
		switch strand {
		case "+":
			// Forward strand: start codon on first exon, stop codon on last exon
			if exonNumberValue == 1 && len(startCodon) > 0 {
				exonObj.StartCodon = &GenomicRange{
					Chrom: startCodon[0].Chrom,
					Start: startCodon[0].Start,
					End:   startCodon[0].End,
				}
			}
			if exonNumberValue == len(exonRecords) && len(stopCodon) > 0 {
				exonObj.StopCodon = &GenomicRange{
					Chrom: stopCodon[0].Chrom,
					Start: stopCodon[0].Start,
					End:   stopCodon[0].End,
				}
			}
		case "-":
			// Reverse strand: start codon on last exon, stop codon on first exon
			if exonNumberValue == len(exonRecords) && len(startCodon) > 0 {
				exonObj.StartCodon = &GenomicRange{
					Chrom: startCodon[0].Chrom,
					Start: startCodon[0].Start,
					End:   startCodon[0].End,
				}
			}
			if exonNumberValue == 1 && len(stopCodon) > 0 {
				exonObj.StopCodon = &GenomicRange{
					Chrom: stopCodon[0].Chrom,
					Start: stopCodon[0].Start,
					End:   stopCodon[0].End,
				}
			}
		}

		// Add UTRs for this exon
		for _, utrRecord := range utrsByExonNumber[exonNumber] {
			exonObj.UTRs = append(exonObj.UTRs, GenomicRange{
				Chrom: utrRecord.Chrom,
				Start: utrRecord.Start,
				End:   utrRecord.End,
			})
		}

		// Add CDSs for this exon
		for _, cdsRecord := range cdsByExonNumber[exonNumber] {
			exonObj.CDSs = append(exonObj.CDSs, GenomicRange{
				Chrom: cdsRecord.Chrom,
				Start: cdsRecord.Start,
				End:   cdsRecord.End,
			})
		}

		exons = append(exons, exonObj)
	}

	return exons, nil
}

func buildTranscripts(geneRecords []Record, strand string) ([]Transcript, error) {
	var transcripts []Transcript

	// Group records by transcript name
	transcriptsByName := filterByAttribute(geneRecords, "transcript_name")

	// Sort transcript names for deterministic order
	transcriptNames := make([]string, 0, len(transcriptsByName))
	for transcriptName := range transcriptsByName {
		transcriptNames = append(transcriptNames, transcriptName)
	}
	sort.Strings(transcriptNames)

	for _, transcriptName := range transcriptNames {
		transcriptRecords := transcriptsByName[transcriptName]
		// Extract the single transcript feature record
		transcriptFeatureRecords := filterByFeature(transcriptRecords, "transcript")
		if len(transcriptFeatureRecords) != 1 {
			return nil, fmt.Errorf("expected exactly one transcript record, got %d", len(transcriptFeatureRecords))
		}
		transcriptRecord := transcriptFeatureRecords[0]

		// Create Transcript object
		transcriptObj := Transcript{
			Feature: Feature{
				Name: transcriptName,
				ID:   transcriptRecord.Attributes["transcript_id"],
				GenomicRange: GenomicRange{
					Chrom: transcriptRecord.Chrom,
					Start: transcriptRecord.Start,
					End:   transcriptRecord.End,
				},
			},
		}

		// Extract start and stop codons for this transcript
		// Note: codons can span splice junctions, resulting in multiple records
		startCodon := filterByFeature(transcriptRecords, "start_codon")
		stopCodon := filterByFeature(transcriptRecords, "stop_codon")

		// Build all exons for this transcript
		exons, err := buildExons(transcriptRecords, startCodon, stopCodon, strand)
		if err != nil {
			return nil, err
		}
		transcriptObj.Exons = exons

		transcripts = append(transcripts, transcriptObj)
	}

	return transcripts, nil
}

func buildGene(geneName string, geneRecords []Record) (Gene, error) {
	// Extract the single gene feature record
	geneFeatureRecords := filterByFeature(geneRecords, "gene")
	if len(geneFeatureRecords) != 1 {
		return Gene{}, fmt.Errorf("expected exactly one gene record, got %d", len(geneFeatureRecords))
	}
	geneRecord := geneFeatureRecords[0]

	// Create Gene object
	geneObj := Gene{
		Feature: Feature{
			Name: geneName,
			ID:   geneRecord.Attributes["gene_id"],
			GenomicRange: GenomicRange{
				Chrom: geneRecord.Chrom,
				Start: geneRecord.Start,
				End:   geneRecord.End,
			},
		},
		Strand: geneRecord.Strand,
		Type:   geneRecord.Attributes["gene_type"],
	}

	// Build all transcripts for this gene
	transcripts, err := buildTranscripts(geneRecords, geneObj.Strand)
	if err != nil {
		return Gene{}, err
	}
	geneObj.Transcripts = transcripts

	return geneObj, nil
}

func filterByAttribute(records []Record, field string) map[string][]Record {
	var result = make(map[string][]Record)
	for _, record := range records {
		value := record.Attributes[field]
		if value == "" {
			continue
		}
		result[value] = append(result[value], record)
	}

	return result
}

func filterByFeature(records []Record, feature string) []Record {
	var result []Record
	for _, record := range records {
		if record.Feature == feature {
			result = append(result, record)
		}
	}
	return result
}
