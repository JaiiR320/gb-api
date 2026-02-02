# Performance Improvements Plan

This document tracks performance optimization work for gb-api.

## Workflow

- When a feature OR large task (many changes) is done, **suggest a git commit** with a short summary
- After commit is done, mark the feature/task(s) as complete

---

## Feature 1: Buffer Pooling for Decompression

Reduce GC pressure by reusing buffers in the hot decompression path.

- [x] **Feature complete**

### Tasks

- [x] Create `sync.Pool` for decompression buffers in `track/bigdata/decompress.go`
  - Add package-level `var bufferPool = sync.Pool{...}` with 64KB default capacity
  
- [x] Modify `DecompressData()` function (`decompress.go:11-28`)
  - Get buffer from pool at start
  - Use `buf.Reset()` to clear
  - `defer bufferPool.Put(buf)` to return
  - Use `io.Copy(buf, reader)` instead of `io.ReadAll`
  - Copy result out before returning (buffer will be reused)

- [x] Add unit test for buffer pool behavior
  - Verify decompression still works correctly
  - Test concurrent decompression calls

**Commit suggestion:** `perf: add sync.Pool for decompression buffers`

---

## Feature 2: Slice Pre-allocation

Eliminate repeated slice reallocations by pre-calculating capacity.

- [x] **Feature complete**

### Tasks

- [x] Pre-allocate in `ReadDataWithZoom()` (`track/bigdata/reader.go:40`)
  - Calculate total capacity from `leafNodes` before loop
  - Use `make([]T, 0, estimatedCapacity)`

- [x] Pre-allocate in `GetCachedWigData()` (`track/bigdata/bigwig/cache.go:143`)
  - Count total points across `rangeData` first
  - Use `make([]BigWigData, 0, totalPoints)`

- [x] Pre-allocate in `GetCachedBedData()` (`track/bigdata/bigbed/cache.go:112`)
  - Count total points across `rangeData` first
  - Use `make([]BigBedData, 0, totalPoints)`

- [x] Pre-allocate in `BrowserHandler()` (`api/handlers.go:102-104`)
  - Use `make([]TrackResponse, 0, len(request.Tracks))`

**Commit suggestion:** `perf: pre-allocate slice capacity in hot paths`

---

## Feature 3: JSON Response Streaming

Stream JSON directly to ResponseWriter instead of double-buffering.

- [ ] **Feature complete**

### Tasks

- [ ] Modify `TrackHandler()` (`api/helpers.go:67-78`)
  - Replace `json.Marshal()` + `w.Write()` with `json.NewEncoder(w).Encode()`
  - Set headers before encoding
  - Handle encoding errors (note: partial response may be sent)

- [ ] Modify `BrowserHandler()` (`api/handlers.go:110-122`)
  - Replace `json.Marshal()` + `w.Write()` with `json.NewEncoder(w).Encode()`
  - Set headers before encoding

- [ ] Update error handling approach
  - Document that streaming means partial responses on encode failure
  - Consider if this tradeoff is acceptable (it usually is)

**Commit suggestion:** `perf: stream JSON responses directly to reduce memory`

---

## Summary

| Feature | Files Modified | Impact |
|---------|----------------|--------|
| Buffer Pooling | `decompress.go` | High - reduces GC in hot path |
| Slice Pre-alloc | `reader.go`, `bigwig/cache.go`, `bigbed/cache.go`, `handlers.go` | Medium - fewer allocations |
| JSON Streaming | `helpers.go`, `handlers.go` | Medium - halves response memory |
