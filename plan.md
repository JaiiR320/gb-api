# Plan: Fix Race Condition in BigWig Cache

## Overview

Fix the race condition in `track/bigdata/bigwig/cache.go` where multiple goroutines write to a shared `erra` variable without synchronization. Replace shared error variable with an error channel pattern.

## Workflow

- When a feature OR large task (many changes) is done, **suggest a git commit** with a short summary
- After commit is done, mark the feature/task(s) as complete

---

## Feature 1: Fix Race Condition in cache.go

- [x] **Task 1.1**: Add error channel
  - Create buffered error channel with capacity 1 (only need first error)
  - `errchan := make(chan error, 1)`

- [x] **Task 1.2**: Replace shared error variable in goroutines
  - Remove `var erra error`
  - Use non-blocking send to capture first error only:
    ```go
    select {
    case errchan <- err:
    default:
    }
    ```

- [x] **Task 1.3**: Update error collection after wait
  - Close error channel after `wg.Wait()`
  - Check for error with `if err, ok := <-errchan; ok { return nil, err }`
  - Remove `erra` from final return

**Commit suggestion**: `fix: resolve race condition in bigwig cache using error channel`

---

## Feature 2: Verify Changes

- [ ] **Task 2.1**: Run tests for bigwig package
  - `go test ./track/bigdata/bigwig/...`

- [ ] **Task 2.2**: Run full test suite
  - `go test ./...`

- [ ] **Task 2.3**: Run race detector
  - `go test -race ./track/bigdata/bigwig/...`

**Commit suggestion**: None (verification only)

---

## Files Changed

| File | Change |
|------|--------|
| `track/bigdata/bigwig/cache.go` | Replace shared error variable with error channel |

---

## Out of Scope

- `track/bigdata/rtree.go` - Uses channel with struct, no race condition
- Adding `context.Context` throughout - Not needed for this fix
- Adding external dependencies - Using stdlib only
