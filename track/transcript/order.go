package transcript

import (
	"sort"
)

// TranscriptLayout represents a transcript with its assigned row
type TranscriptLayout struct {
	ID  string
	Row int
}

// transcriptItem is used internally for the packing algorithm
type transcriptItem struct {
	id          string
	start       int
	end         int
	layoutStart int
	layoutEnd   int
}

// PackTranscripts assigns row indices to transcripts using UCSC-style interval packing.
// It minimizes vertical height while avoiding horizontal overlap.
//
// Parameters:
//   - transcripts: slice of LegacyTranscript to layout
//   - paddingBp: horizontal padding in base pairs between transcripts (default: 100)
//
// Returns:
//   - map from transcript ID to row index (0 = top row)
func PackTranscripts(transcripts []LegacyTranscript, paddingBp int) map[string]int {
	if len(transcripts) == 0 {
		return make(map[string]int)
	}

	if paddingBp <= 0 {
		paddingBp = 100 // Default padding in base pairs
	}

	// Prepare items with layout intervals
	items := make([]transcriptItem, 0, len(transcripts))
	for _, tx := range transcripts {
		items = append(items, transcriptItem{
			id:          tx.ID,
			start:       tx.Coordinates.Start,
			end:         tx.Coordinates.End,
			layoutStart: tx.Coordinates.Start - paddingBp,
			layoutEnd:   tx.Coordinates.End + paddingBp,
		})
	}

	// Sort deterministically: layoutStart, then layoutEnd, then id
	sort.Slice(items, func(i, j int) bool {
		if items[i].layoutStart != items[j].layoutStart {
			return items[i].layoutStart < items[j].layoutStart
		}
		if items[i].layoutEnd != items[j].layoutEnd {
			return items[i].layoutEnd < items[j].layoutEnd
		}
		return items[i].id < items[j].id
	})

	// Greedy first-fit interval packing
	rowsRightEnd := make([]int, 0) // Track the rightmost end coordinate for each row
	result := make(map[string]int, len(items))

	for _, item := range items {
		placedRow := -1

		// Find the lowest-index row that fits (no overlap)
		for r := 0; r < len(rowsRightEnd); r++ {
			if item.layoutStart >= rowsRightEnd[r] {
				placedRow = r
				rowsRightEnd[r] = item.layoutEnd
				break
			}
		}

		// If no row fits, create a new row
		if placedRow == -1 {
			placedRow = len(rowsRightEnd)
			rowsRightEnd = append(rowsRightEnd, item.layoutEnd)
		}

		result[item.id] = placedRow
	}

	return result
}

// PackTranscriptsWithLayout is a convenience function that returns a slice of TranscriptLayout
// instead of a map, preserving the original order of transcripts.
func PackTranscriptsWithLayout(transcripts []LegacyTranscript, paddingBp int) []TranscriptLayout {
	rowMap := PackTranscripts(transcripts, paddingBp)

	result := make([]TranscriptLayout, 0, len(transcripts))
	for _, tx := range transcripts {
		result = append(result, TranscriptLayout{
			ID:  tx.ID,
			Row: rowMap[tx.ID],
		})
	}

	return result
}

// GetTotalRows returns the total number of rows needed for the given layout
func GetTotalRows(layout map[string]int) int {
	if len(layout) == 0 {
		return 0
	}

	maxRow := 0
	for _, row := range layout {
		if row > maxRow {
			maxRow = row
		}
	}
	return maxRow + 1
}
