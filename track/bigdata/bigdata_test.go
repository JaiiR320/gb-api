package bigdata

import "testing"

func TestSelectZoomLevel_NoZoomLevels(t *testing.T) {
	b := &BigData{ZoomLevels: []ZoomLevelHeader{}}

	zoomIdx := b.SelectZoomLevel(0, 10000, 1000)
	if zoomIdx != -1 {
		t.Errorf("Expected -1 for no zoom levels, got %d", zoomIdx)
	}
}

func TestSelectZoomLevel_InvalidWidth(t *testing.T) {
	b := &BigData{
		ZoomLevels: []ZoomLevelHeader{
			{ReductionLevel: 10, IndexOffset: 1000},
		},
	}

	// preRenderedWidth = 0
	zoomIdx := b.SelectZoomLevel(0, 10000, 0)
	if zoomIdx != -1 {
		t.Errorf("Expected -1 for preRenderedWidth=0, got %d", zoomIdx)
	}

	// preRenderedWidth < 0
	zoomIdx = b.SelectZoomLevel(0, 10000, -10)
	if zoomIdx != -1 {
		t.Errorf("Expected -1 for negative preRenderedWidth, got %d", zoomIdx)
	}
}

func TestSelectZoomLevel_BelowThreshold(t *testing.T) {
	b := &BigData{
		ZoomLevels: []ZoomLevelHeader{
			{ReductionLevel: 10, IndexOffset: 1000},
		},
	}

	// basesPerPixel = 100 / 100 = 1 < 2 (threshold)
	zoomIdx := b.SelectZoomLevel(0, 100, 100)
	if zoomIdx != -1 {
		t.Errorf("Expected -1 for below threshold, got %d", zoomIdx)
	}
}

func TestSelectZoomLevel_SelectsBestZoom(t *testing.T) {
	b := &BigData{
		ZoomLevels: []ZoomLevelHeader{
			{Index: 0, ReductionLevel: 10, IndexOffset: 1000},
			{Index: 1, ReductionLevel: 100, IndexOffset: 2000},
			{Index: 2, ReductionLevel: 1000, IndexOffset: 3000},
		},
	}

	// basesPerPixel = 10,000,000 / 1000 = 10,000
	// Should select zoom level 2 (reduction=1000) - highest that doesn't exceed 10,000
	zoomIdx := b.SelectZoomLevel(0, 10_000_000, 1000)
	if zoomIdx != 2 {
		t.Errorf("Expected zoom level 2, got %d", zoomIdx)
	}
}

func TestSelectZoomLevel_AllTooCoarse(t *testing.T) {
	b := &BigData{
		ZoomLevels: []ZoomLevelHeader{
			{Index: 0, ReductionLevel: 100, IndexOffset: 1000},
			{Index: 1, ReductionLevel: 1000, IndexOffset: 2000},
		},
	}

	// basesPerPixel = 1000 / 100 = 10
	// All zoom levels have reduction > 10, so use full resolution
	zoomIdx := b.SelectZoomLevel(0, 1000, 100)
	if zoomIdx != -1 {
		t.Errorf("Expected -1 when all zooms too coarse, got %d", zoomIdx)
	}
}

func TestSelectZoomLevel_EdgeCase(t *testing.T) {
	b := &BigData{
		ZoomLevels: []ZoomLevelHeader{
			{Index: 0, ReductionLevel: 10, IndexOffset: 1000},
			{Index: 1, ReductionLevel: 100, IndexOffset: 2000},
			{Index: 2, ReductionLevel: 1000, IndexOffset: 3000},
			{Index: 3, ReductionLevel: 10000, IndexOffset: 4000},
		},
	}

	// basesPerPixel = 100,000,000 / 800 = 125,000
	// Should select zoom level 3 (reduction=10,000) - highest that doesn't exceed 125,000
	zoomIdx := b.SelectZoomLevel(0, 100_000_000, 800)
	if zoomIdx != 3 {
		t.Errorf("Expected zoom level 3, got %d", zoomIdx)
	}
}

func TestSelectZoomLevel_ExactMatch(t *testing.T) {
	b := &BigData{
		ZoomLevels: []ZoomLevelHeader{
			{Index: 0, ReductionLevel: 10, IndexOffset: 1000},
			{Index: 1, ReductionLevel: 100, IndexOffset: 2000},
			{Index: 2, ReductionLevel: 1000, IndexOffset: 3000},
		},
	}

	// basesPerPixel = 100,000 / 100 = 1000 (exact match with zoom level 2)
	zoomIdx := b.SelectZoomLevel(0, 100_000, 100)
	if zoomIdx != 2 {
		t.Errorf("Expected zoom level 2 for exact match, got %d", zoomIdx)
	}
}
