# AGENTS.md

This file provides guidance to AI coding agents working in this repository.

## Build and Run Commands

```bash
# Build the server binary
go build -o main .

# Run server (listens on port 8080)
./main

# Run all tests
go test ./...

# Run tests for a specific package
go test ./api
go test ./cache
go test ./track/transcript

# Run a single test by name
go test ./api -run TestBigWigHandler
go test ./cache -run TestFindUncachedRanges

# Run tests with verbose output
go test -v ./...

# Deploy to Fly.io
fly deploy
```

## Architecture Overview

This is a Go HTTP API server for a genome browser, serving genomic data tracks via HTTP endpoints.

### Directory Structure
```
gb-api/
├── main.go                 # Entry point, route registration
├── api/                    # HTTP handlers and request/response types
│   ├── handlers.go         # Endpoint handlers (/bigwig, /bigbed, /transcript, /browser)
│   ├── helpers.go          # Generic TrackHandler wrapper
│   ├── model.go            # Request/response structs
│   └── middleware/         # CORS and rate limiting
├── cache/                  # Generic LRU cache and range caching
├── track/                  # Track data processing
│   ├── bigdata/            # Shared BigWig/BigBed infrastructure (bigwig/, bigbed/)
│   └── transcript/         # GTF parsing for gene/transcript data
└── utils/                  # Binary parsing utilities
```

### Key Patterns
- Track handlers use generics: `TrackHandler[T TrackRequest](...)`
- `/browser` endpoint aggregates multiple tracks in parallel using goroutines
- BigWig/BigBed use LRU caching with range-aware merging
- Transcript data reads from local GTF file: `./track/transcript/data/v40/sorted.gtf.gz`

## Code Style Guidelines

### Import Organization
Imports grouped: 1) Standard library, 2) Internal packages (gb-api/...), 3) External deps

```go
import (
    "encoding/json"
    "net/http"

    "gb-api/track/bigdata/bigwig"

    lru "github.com/hashicorp/golang-lru/v2"
)
```

### Naming Conventions
- **Files**: lowercase with underscores (e.g., `rangecache.go`, `zoom_decoder.go`)
- **Packages**: short, lowercase, no underscores (e.g., `bigwig`, `transcript`)
- **Types**: PascalCase (e.g., `BigWigData`, `TrackResponse`)
- **Functions**: PascalCase for exported, camelCase for internal
- **Constants**: SCREAMING_SNAKE_CASE for magic numbers, PascalCase for typed constants
- **Variables**: camelCase (e.g., `basesPerPixel`, `zoomIdx`)

```go
const BIGWIG_MAGIC_LTH = 0x888FFC26  // Magic number constant
const defaultPaddingBp = 100         // Local constant
```

### Struct Definitions
- Use JSON tags for API-facing structs
- Use `omitempty` for optional fields
- Group related fields together

```go
type BigWigData struct {
    Chr   string  `json:"chr"`
    Start int32   `json:"start"`
    End   int32   `json:"end"`
    Value float32 `json:"value"`
}
```

### Error Handling
- Wrap errors with context using `fmt.Errorf("..., %w", err)`
- Return errors early, avoid deep nesting
- Log errors with structured logging (`slog`)

### HTTP Handlers
- Generate request UUID for tracing
- Use structured logging with `slog`
- Set Content-Type header before writing response

### Testing Patterns
- Use table-driven tests with descriptive names
- Use `t.Run()` for subtests
- Use `t.Helper()` in helper functions
- Use `t.Skip()` when test data is unavailable

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {name: "descriptive case name", input: ..., expected: ...},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
                return
            }
            if got != tt.expected {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### Concurrency
- Use channels for goroutine communication
- Use `sync.RWMutex` for thread-safe caches
- Collect results via buffered channels

### Generics
This codebase uses Go generics for type-safe caching and handlers.
See `cache/cache.go` and `api/helpers.go` for examples.

## Dependencies
- `github.com/hashicorp/golang-lru/v2` - LRU cache
- `github.com/brentp/bix` - Tabix-indexed file access
- `golang.org/x/time/rate` - Rate limiting
