package bigwig

import (
	"math"
)

// PrerenderedBin represents a single bin in the prerendered output
type PrerenderedBin struct {
	X   int     `json:"x"`   // Bin index (0-based)
	Max float32 `json:"max"` // Maximum value in this bin
	Min float32 `json:"min"` // Minimum value in this bin
}

// ResampleToWidth resamples BigWig data to exactly targetWidth bins.
// Each bin contains the max and min values found in that genomic region.
// The genomic range is split evenly by targetWidth.
func ResampleToWidth(data []BigWigData, targetWidth int) []PrerenderedBin {
	if len(data) == 0 || targetWidth <= 0 {
		return []PrerenderedBin{}
	}

	// Calculate the genomic range from data points
	start := data[0].Start
	end := data[len(data)-1].End
	totalRange := float64(end - start)

	if totalRange <= 0 {
		return []PrerenderedBin{}
	}

	// Calculate bin size in genomic coordinates
	binSize := totalRange / float64(targetWidth)

	// Initialize result bins - use pointers to track if bin has data
	type binData struct {
		hasData bool
		max     float32
		min     float32
	}
	bins := make([]binData, targetWidth)

	// Single pass through data points - O(n) instead of O(n×m)
	for _, point := range data {
		// Calculate which bins this point overlaps with
		// Find first bin that this point touches
		firstBin := int(float64(point.Start-start) / binSize)
		if firstBin < 0 {
			firstBin = 0
		}
		if firstBin >= targetWidth {
			firstBin = targetWidth - 1
		}

		// Find last bin that this point touches
		lastBin := int(float64(point.End-start) / binSize)
		if lastBin < 0 {
			lastBin = 0
		}
		if lastBin >= targetWidth {
			lastBin = targetWidth - 1
		}

		// Update all bins that this point overlaps
		for binIdx := firstBin; binIdx <= lastBin; binIdx++ {
			if !bins[binIdx].hasData {
				bins[binIdx].hasData = true
				bins[binIdx].max = point.Value
				bins[binIdx].min = point.Value
			} else {
				if point.Value > bins[binIdx].max {
					bins[binIdx].max = point.Value
				}
				if point.Value < bins[binIdx].min {
					bins[binIdx].min = point.Value
				}
			}
		}
	}

	// Build result, filling gaps with nearest neighbor values
	result := make([]PrerenderedBin, targetWidth)
	var lastValidValue float32
	hasLastValid := false

	for i := range targetWidth {
		if bins[i].hasData {
			result[i] = PrerenderedBin{
				X:   i,
				Max: bins[i].max,
				Min: bins[i].min,
			}
			lastValidValue = bins[i].max // Use max as representative value
			hasLastValid = true
		} else {
			// Use last valid value, or find next valid value if no previous exists
			fillValue := lastValidValue
			if !hasLastValid {
				// Look forward to find first valid bin
				for j := i + 1; j < targetWidth; j++ {
					if bins[j].hasData {
						fillValue = bins[j].max
						break
					}
				}
			}
			result[i] = PrerenderedBin{
				X:   i,
				Max: fillValue,
				Min: fillValue,
			}
		}
	}

	return result
}

// findNearestValue finds the value from the nearest data point to the given range
func findNearestValue(data []BigWigData, binStart, binEnd int32) float32 {
	if len(data) == 0 {
		return 0
	}

	binMid := (binStart + binEnd) / 2

	// Find closest point to bin midpoint
	minDist := int32(math.MaxInt32)
	var closestValue float32

	for _, point := range data {
		pointMid := (point.Start + point.End) / 2
		dist := abs(pointMid - binMid)
		if dist < minDist {
			minDist = dist
			closestValue = point.Value
		}
	}

	return closestValue
}

func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
