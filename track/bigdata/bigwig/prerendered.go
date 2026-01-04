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

	result := make([]PrerenderedBin, targetWidth)

	for i := 0; i < targetWidth; i++ {
		binStart := start + int32(float64(i)*binSize)
		binEnd := start + int32(float64(i+1)*binSize)

		// Find all data points that overlap with this bin
		var values []float32
		for _, point := range data {
			// Check if point overlaps with bin
			if point.End > binStart && point.Start < binEnd {
				values = append(values, point.Value)
			}
		}

		// Calculate max and min for this bin
		var maxVal, minVal float32
		if len(values) > 0 {
			maxVal = values[0]
			minVal = values[0]
			for _, v := range values {
				if v > maxVal {
					maxVal = v
				}
				if v < minVal {
					minVal = v
				}
			}
		} else {
			// If no values in bin, use nearest value for both max and min
			nearestVal := findNearestValue(data, binStart, binEnd)
			maxVal = nearestVal
			minVal = nearestVal
		}

		result[i] = PrerenderedBin{
			X:   i,
			Max: maxVal,
			Min: minVal,
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
