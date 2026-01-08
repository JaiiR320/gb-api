# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Build
go build -o main .

# Run server (listens on port 8080)
./main

# Run all tests
go test ./...

# Run tests for a specific package
go test ./api
go test ./cache
go test ./track/transcript

# Run a single test
go test ./api -run TestBigWigHandler

# Deploy to Fly.io
fly deploy
```

## Architecture Overview

This is a Go HTTP API server for a genome browser, serving genomic data tracks (BigWig, BigBed, and transcript data) via HTTP endpoints.

### API Layer (`api/`)
- `handlers.go` - HTTP handlers for `/bigwig`, `/bigbed`, `/transcript`, and `/browser` endpoints
- `model.go` - Request/response types and track configuration structs
- The `/browser` endpoint aggregates multiple track requests in parallel using goroutines

### Track Data (`track/`)
- `track/bigdata/` - Shared BigWig/BigBed parsing infrastructure (headers, R-tree, decompression)
  - `bigwig/` - BigWig file reading with LRU caching and pre-rendering/resampling
  - `bigbed/` - BigBed file reading with header caching and parsers (e.g., CCRE)
- `track/transcript/` - GTF file parsing for gene/transcript data with row packing layout

### Caching (`cache/`)
- Generic thread-safe LRU cache wrapper (`Cache[T]`)
- `RangeCache` for genomic range queries - finds uncached ranges to fetch

### Key Patterns
- BigWig uses `GetCachedWigData()` which caches data by URL+chrom, merging overlapping ranges
- BigBed caches parsed headers per URL
- Transcript data reads from local GTF file (`./track/transcript/data/v40/sorted.gtf.gz`)
- Track handlers use generics: `TrackHandler[T TrackRequest](...)` with type-specific processing functions
