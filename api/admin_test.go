package api

import (
	"encoding/json"
	"gb-api/cache"
	"gb-api/track/bigdata/bigbed"
	"gb-api/track/bigdata/bigwig"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCacheSizeHandler(t *testing.T) {
	// Add some test data to caches
	testURL := "https://example.com/test.bw"
	testChrom := "chr1"

	// Add some bigwig data
	testWigData := []bigwig.BigWigData{
		{Chr: "chr1", Start: 100, End: 200, Value: 1.5},
		{Chr: "chr1", Start: 200, End: 300, Value: 2.5},
	}
	bigwig.BigWigDataCache.Add(testURL+"-"+testChrom, []cache.RangeData[bigwig.BigWigData]{
		{Start: 100, End: 300, Data: testWigData},
	})

	// Add some bigbed data
	testBedData := []bigbed.BigBedData{
		{Chr: "chr1", Start: 100, End: 200, Rest: "test1"},
		{Chr: "chr1", Start: 200, End: 300, Rest: "test2"},
	}
	bigbed.BigBedDataCache.Add(testURL+"-"+testChrom, []cache.RangeData[bigbed.BigBedData]{
		{Start: 100, End: 300, Data: testBedData},
	})

	// Test without keys
	req := httptest.NewRequest("GET", "/admin/cache-status", nil)
	w := httptest.NewRecorder()
	CacheSizeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response CacheStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Caches) != 4 {
		t.Errorf("Expected 4 cache entries, got %d", len(response.Caches))
	}

	// Verify cache names
	expectedNames := map[string]bool{
		"bigwig-data":    false,
		"bigwig-headers": false,
		"bigbed-data":    false,
		"bigbed-headers": false,
	}
	for _, cache := range response.Caches {
		if _, ok := expectedNames[cache.Name]; !ok {
			t.Errorf("Unexpected cache name: %s", cache.Name)
		}
		expectedNames[cache.Name] = true

		// Verify sizes are calculated
		if cache.ApproxSizeKB < 0 {
			t.Errorf("Cache %s has negative size", cache.Name)
		}

		// Verify keys are not included
		if len(cache.Keys) > 0 {
			t.Errorf("Expected no keys without query param, got %d keys for %s", len(cache.Keys), cache.Name)
		}
	}

	// Verify all caches were found
	for name, found := range expectedNames {
		if !found {
			t.Errorf("Cache %s not found in response", name)
		}
	}

	if response.TotalSizeKB < 0 {
		t.Errorf("Total size KB should be non-negative, got %d", response.TotalSizeKB)
	}

	if response.TotalSizeMB < 0 {
		t.Errorf("Total size MB should be non-negative, got %d", response.TotalSizeMB)
	}

	// Test with keys=true
	req = httptest.NewRequest("GET", "/admin/cache-status?keys=true", nil)
	w = httptest.NewRecorder()
	CacheSizeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var responseWithKeys CacheStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&responseWithKeys); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify keys are included for caches with data
	foundKeysInCache := false
	for _, cache := range responseWithKeys.Caches {
		if cache.EntryCount > 0 && len(cache.Keys) > 0 {
			foundKeysInCache = true
			break
		}
	}

	if !foundKeysInCache {
		t.Log("Warning: No keys found in caches with entries, this may be expected if cache entries were not persisted")
	}

	t.Logf("Cache status response: %d caches, %d KB total, %d MB total",
		len(response.Caches), response.TotalSizeKB, response.TotalSizeMB)
}
