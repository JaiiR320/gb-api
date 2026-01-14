# Production Readiness Plan

This document tracks API improvements to bring gb-api from prototype to production.

## Segment 1: Quick Wins (Current)

- [ ] 1.1 Add `/health` endpoint
- [ ] 1.2a Replace debug prints in bigwig cache with slog
- [ ] 1.2b Replace debug prints in bigbed cache with slog
- [ ] 1.2c Delete unused `PrintRecord` function
- [ ] 1.2d Replace main.go prints with slog
- [ ] 1.3 Remove hardcoded `X-Cache-Status` header
- [ ] 1.4 Add `X-Request-ID` response header

## Segment 2: Error Handling

- [ ] 2.1 Create standardized error response struct
- [ ] 2.2 Add input validation (start/end, chrom, URL)
- [ ] 2.3 Add rate limit headers (X-RateLimit-*)

## Segment 3: Reliability

- [ ] 3.1 Add request timeouts with context propagation
- [ ] 3.2 Implement graceful shutdown (SIGTERM/SIGINT)
- [ ] 3.3 Add request body size limits

## Segment 4: Configuration

- [ ] 4.1 Environment-based config (port, cache size)
- [ ] 4.2 Add security headers (X-Content-Type-Options, etc.)

## Segment 5: Documentation & Versioning

- [ ] 5.1 Add OpenAPI/Swagger spec
- [ ] 5.2 Add API versioning (/v1/ prefix)

## Segment 6: Future Enhancements

- [ ] 6.1 Batch region requests
- [ ] 6.2 Response compression (gzip)
