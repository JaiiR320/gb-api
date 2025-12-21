# Code Review Issues - API Directory

Generated: 2025-12-12

---

## 1. Critical Issues

### 1.1 Missing `return` Statement After Error (BUG - handlers.go:51-53) ✅ FIXED

**File:** `handlers.go`

```go
data, err := transcript.GetTranscripts(request.Chrom, request.Start, request.End)
if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    // MISSING: return
}
```

**Problem:** When `GetTranscripts` returns an error, the handler sends an HTTP error response but **continues executing**. This will attempt to encode a nil/empty `data` response after already writing an error, causing either a panic or corrupted response.

**Impact:** Runtime failures, inconsistent API responses, potential panic.

**Fix:** Add `return` after `http.Error()`.

**Status:** FIXED - Return statement is now present on line 55.

---

### 1.2 Missing `return` After Decode Error (BUG - handlers.go:71-74) ✅ FIXED

**File:** `handlers.go`

```go
err := json.NewDecoder(r.Body).Decode(&request)
if err != nil {
    fmt.Printf("ERROR: Failed to decode request: %v\n", err)
    http.Error(w, err.Error(), http.StatusBadRequest)
    // MISSING: return
}
```

**Problem:** Same issue as above - execution continues after error, leading to undefined behavior when `request` has zero values.

**Status:** FIXED - Return statement is now present on line 78.

---

### 1.3 Incorrect HTTP Status Codes (handlers.go:23, 33-34) ✅ FIXED

**File:** `handlers.go`

```go
data, err := bigwig.ReadBigWig(...)
if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)  // Should distinguish error types
}
```

**Problem:** All errors return `400 Bad Request`, even when:
- The external URL is unreachable (should be `502 Bad Gateway` or `500 Internal Server Error`)
- Internal processing fails (should be `500 Internal Server Error`)
- Resource not found (should be `404 Not Found`)

**Impact:** API consumers cannot distinguish between client errors (bad input) and server errors (internal failures), making debugging difficult.

**Status:** FIXED - Data fetching and encoding errors now return 500 Internal Server Error:
- `bigwig.ReadBigWig` errors → 500
- `transcript.GetTranscripts` errors → 500
- JSON marshal/encode errors → 500
- JSON decode errors → 400 (correct - client sent invalid JSON)

---

### 1.4 Silent Error Handling - Ignored Marshal Error (handlers.go:114) ✅ FIXED

**File:** `handlers.go`

```go
d, _ := json.Marshal(data)  // Error is silently ignored
```

**Problem:** If marshaling fails, `d` will be nil/empty, and the response will contain no data without any indication of failure.

**Status:** FIXED - Now properly captures and handles marshal errors in BrowserHandler goroutine, sending error in TrackResponse.

---

## 2. Security Concerns

### 2.1 CORS Configuration Too Permissive (middleware.go:9) ✅ RESOLVED - INTENTIONAL

**File:** `middleware.go`

```go
w.Header().Set("Access-Control-Allow-Origin", "*")
```

**Problem:** Allowing all origins (`*`) can expose your API to:
- Cross-site request forgery (CSRF) attacks
- Unauthorized access from malicious websites
- Data exfiltration

**Recommendation:** Specify allowed origins explicitly, or at minimum, make this configurable via environment variables:

```go
allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
if allowedOrigin == "" {
    allowedOrigin = "https://yourdomain.com"
}
w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
```

**Status:** RESOLVED - Intentional design choice documented in code:
- This is a public API serving genomic data with no authentication
- Wildcard `*` allows third-party developers to build tools using the API
- Enables access from Vercel-hosted frontend and other origins
- Documented with comment explaining rationale and future considerations
- If authentication or sensitive data is added, CORS should be restricted

---

### 2.2 Unrestricted External URL Access (handlers.go:21, 97) ✅ RESOLVED - INTENTIONAL

**File:** `handlers.go`

```go
data, err := bigwig.ReadBigWig(request.URL, ...)  // User-provided URL
```

**Problem:** The API accepts arbitrary URLs from clients. This creates:
- **SSRF (Server-Side Request Forgery)** risk - attackers could access internal resources
- **Resource exhaustion** - malicious users could point to extremely large files
- **Information disclosure** - probing internal network topology

**Recommendations:**
1. Implement URL allowlisting (whitelist of allowed domains)
2. Validate URL schemes (only `https://`)
3. Block private/internal IP ranges
4. Set request timeouts and size limits

**Status:** RESOLVED - Intentional design choice documented in code:
- API designed for maximum flexibility - users can visualize BigWig data from any source
- Supports public databases, company internal servers, and local files (localhost)
- Core feature for experimentation and personal data exploration
- BigWig format validation (header check) provides protection against accessing non-BigWig resources
- Restricting URLs would break the intended use case
- Documented with detailed comment explaining rationale and supported URL types

---

### 2.3 Error Messages Expose Internal Details

**File:** `handlers.go`

```go
http.Error(w, err.Error(), http.StatusBadRequest)
```

**Problem:** Raw error messages may leak sensitive information about:
- Internal file paths
- Database structures
- Third-party service URLs

**Recommendation:** Log detailed errors server-side, return generic messages to clients:

```go
log.Printf("Error processing request: %v", err)
http.Error(w, "Failed to process request", http.StatusInternalServerError)
```

---

## 3. Significant Improvements

### 3.1 No HTTP Method Validation ✅ FIXED

**File:** `handlers.go`

**Problem:** Handlers do not verify the HTTP method. While CORS middleware handles OPTIONS, the handlers accept GET, DELETE, PUT, etc., when they should only accept POST.

**Fix:**
```go
if r.Method != http.MethodPost {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
}
```

**Status:** FIXED - All handlers now validate HTTP method:
- `handleTrackRequest` validates method for BigWigHandler and TranscriptHandler
- `BrowserHandler` validates method directly
- All endpoints now return 405 Method Not Allowed for non-POST requests (except OPTIONS, which is handled by CORS middleware)

---

### 3.2 No Request Body Size Limits

**Problem:** No limits on request body size allows memory exhaustion attacks.

**Fix:**
```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit
```

---

### 3.3 No Input Validation

**File:** `model.go`

**Problem:** Request structs have no validation:
- `Start` could be negative
- `End` could be less than `Start`
- `Chrom` could be empty or invalid
- `URL` is not validated

**Recommendation:** Add validation methods:

```go
func (r *BigWigRequest) Validate() error {
    if r.Chrom == "" {
        return errors.New("chrom is required")
    }
    if r.Start < 0 || r.End < 0 {
        return errors.New("start and end must be non-negative")
    }
    if r.End <= r.Start {
        return errors.New("end must be greater than start")
    }
    if r.URL == "" {
        return errors.New("url is required")
    }
    return nil
}
```

---

### 3.4 Inconsistent Type Usage for Coordinates ✅ FIXED

**File:** `model.go`

| Struct | Start Type | End Type |
|--------|-----------|----------|
| `BigWigRequest` | `int32` | `int32` |
| `TranscriptRequest` | `int` | `int` |
| `BrowserRequest` | `int` | `int` |

**Problem:** Inconsistent types cause unnecessary type conversions (see line 97: `int32(request.Start)`) and potential overflow issues.

**Recommendation:** Standardize on a single type (preferably `int64` for genomic coordinates to handle large values safely).

**Status:** FIXED - All coordinate types standardized to `int`:
- `BigWigRequest.Start` and `BigWigRequest.End` changed from `int32` to `int`
- Type conversions removed from `BrowserHandler`
- All request structs now use consistent `int` type for coordinates

---

### 3.5 Missing JSON Tags on BrowserRequest ✅ FIXED

**File:** `model.go`

```go
type BrowserRequest struct {
    Chrom  string    // Missing `json:"chrom"`
    Start  int       // Missing `json:"start"`
    End    int       // Missing `json:"end"`
    Tracks []Track
}
```

**Problem:** Without JSON tags, the JSON decoder relies on case-insensitive matching, which works but is inconsistent with other structs and less explicit.

**Status:** FIXED - All fields in BrowserRequest now have explicit JSON tags:
- `Chrom` → `json:"chrom"`
- `Start` → `json:"start"`
- `End` → `json:"end"`
- Consistent with other request structs

---

### 3.6 Console Logging Instead of Proper Logging ✅ FIXED

**File:** `handlers.go`

```go
fmt.Println("received bigwig request")
fmt.Printf("ERROR: Failed to decode request: %v\n", err)
```

**Problems:**
- No log levels (debug, info, error)
- No structured logging (JSON for production)
- No timestamps
- Cannot be disabled/configured
- Typo: "recieved" should be "received" (line 41)

**Recommendation:** Use a structured logging library like `log/slog` (stdlib) or `zerolog`:

```go
slog.Info("received bigwig request", "chrom", request.Chrom, "start", request.Start)
```

**Status:** FIXED - Implemented structured logging with `log/slog`:
- Created request IDs using crypto/rand for tracking individual requests
- Created loggers with default requestID field using `slog.With()`
- Added Info and Error logs with structured fields throughout all handlers
- Logs include contextual information (error, chrom, start, end, url, method, etc.)
- Track-specific logging in getTrackData goroutines

---

### 3.7 Unbuffered Channel in BrowserHandler

**File:** `handlers.go`

```go
var results = make(chan TrackResponse, len(request.Tracks))
```

This is actually correctly buffered, but there is no **timeout mechanism**. If a track processing goroutine hangs, the handler will block forever.

**Recommendation:** Add context with timeout:

```go
ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
defer cancel()
```

---

### 3.8 Type Assertion Without Safety Check

**File:** `handlers.go`

```go
cfg := config.(BigWigConfig)  // Will panic if assertion fails
```

**Problem:** If `GetConfig()` returns an unexpected type, this will panic.

**Fix:**
```go
cfg, ok := config.(BigWigConfig)
if !ok {
    // Handle error
}
```

---

## 4. Testing Issues

### 4.1 Tests Are Integration Tests, Not Unit Tests

**File:** `handler_test.go`

**Problem:** All tests make HTTP calls to `localhost:8080`:
- Requires a running server
- Tests external systems (network, files)
- Cannot run in CI/CD without infrastructure
- Tests are not isolated or repeatable

**Recommendation:** Use `httptest` for true unit tests:

```go
func TestBigWigHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodPost, "/bigwig", strings.NewReader(`{...}`))
    w := httptest.NewRecorder()

    BigWigHandler(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
}
```

---

### 4.2 Tests Use Relative Paths

**File:** `handler_test.go`

```go
err := utils.ReadFromJSON("test/request/bigWigRequest.json", &request)
```

**Problem:** Relative paths are fragile - tests will fail if run from a different directory.

---

### 4.3 No Error Assertions in Tests

```go
if err != nil {
    t.Error(err.Error())
}
```

**Problems:**
- Tests continue after errors (use `t.Fatal()` for critical errors)
- No assertions on response content
- No negative test cases (invalid input, error conditions)

---

### 4.4 Missing Test Coverage

No tests for:
- Invalid JSON input
- Missing required fields
- Invalid coordinates (negative, end < start)
- Error responses
- Middleware behavior
- Concurrent requests

---

## 5. Code Quality Enhancements

### 5.1 Dead Code / Unused Struct

**File:** `model.go`

```go
type BigBedRequest struct { ... }  // Defined but never used
```

---

### 5.2 Response Encoding Error Handling Inconsistency ✅ FIXED

**File:** `handlers.go`

| Handler | Error Handling |
|---------|---------------|
| `BigWigHandler` | Returns `http.Error` (but too late - headers sent) |
| `TranscriptHandler` | Just prints error |
| `BrowserHandler` | Just prints error |

**Problem:** Inconsistent and ineffective - once `w.Header().Set()` or `w.WriteHeader()` is called, you cannot change the status code.

**Status:** FIXED - All handlers now use consistent error handling:
- Marshal response to bytes first using `json.Marshal()`
- Check for errors before setting any headers
- Only set `Content-Type` and write response after successful marshaling
- Return proper 500 status code for encoding errors (server-side issue, not client error)

---

## Summary of Recommended Actions

### Immediate (Critical Bugs):
1. ~~Add missing `return` statements after error handling in `TranscriptHandler` and `BrowserHandler`~~ ✅ COMPLETED
2. Fix type assertion safety in `BrowserHandler`
3. ~~Handle JSON marshal errors properly~~ ✅ COMPLETED
4. ~~Fix incorrect HTTP status codes~~ ✅ COMPLETED

### Short-term (Security & Reliability):
4. Add input validation for all request types
5. ~~Implement URL allowlisting for BigWig requests~~ ✅ RESOLVED (intentional design for flexibility)
6. ~~Add HTTP method validation~~ ✅ COMPLETED
7. ~~CORS configuration~~ ✅ RESOLVED (intentional public API design)
8. Add request body size limits
9. Standardize error responses with proper status codes

### Medium-term (Testing & Quality):
9. Convert integration tests to unit tests using `httptest`
10. Add comprehensive test coverage including error cases
11. ~~Implement structured logging~~ ✅ COMPLETED
12. ~~Standardize coordinate types across all structs~~ ✅ COMPLETED
13. ~~Add missing JSON tags to BrowserRequest~~ ✅ COMPLETED

### Long-term (Architecture):
13. Add request timeout/context handling
14. Make CORS origins configurable
15. Add health check endpoint
16. Consider adding API versioning (e.g., `/v1/bigwig`)
