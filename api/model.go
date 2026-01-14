package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
)

// APIError represents a standardized error response
type APIError struct {
	Code    string `json:"code"`              // Machine-readable error code
	Message string `json:"message"`           // Human-readable message
	Field   string `json:"field,omitempty"`   // Field that caused the error (for validation)
	Details string `json:"details,omitempty"` // Additional context
}

// ErrorResponse wraps an APIError for JSON responses
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// Common error codes
const (
	ErrCodeMethodNotAllowed  = "METHOD_NOT_ALLOWED"
	ErrCodeInvalidJSON       = "INVALID_JSON"
	ErrCodeValidation        = "VALIDATION_ERROR"
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
)

// NewAPIError creates a new APIError
func NewAPIError(code, message string) APIError {
	return APIError{Code: code, Message: message}
}

// NewValidationError creates a validation error for a specific field
func NewValidationError(field, message string) APIError {
	return APIError{
		Code:    ErrCodeValidation,
		Message: message,
		Field:   field,
	}
}

// chromRegex validates chromosome format (chr1-22, chrX, chrY, chrM, etc.)
var chromRegex = regexp.MustCompile(`^chr([1-9]|1[0-9]|2[0-2]|X|Y|M|MT)$`)

// Validatable interface for request validation
type Validatable interface {
	Validate() *APIError
}

// Track requests
type BigWigRequest struct {
	URL              string `json:"url"`
	Chrom            string `json:"chrom"`
	Start            int    `json:"start"`
	End              int    `json:"end"`
	PreRenderedWidth int    `json:"preRenderedWidth,omitempty"` // Number of points to return
}

// Validate checks BigWigRequest fields
func (r *BigWigRequest) Validate() *APIError {
	if r.URL == "" {
		err := NewValidationError("url", "url is required")
		return &err
	}
	if _, parseErr := url.ParseRequestURI(r.URL); parseErr != nil {
		err := NewValidationError("url", fmt.Sprintf("invalid url: %s", parseErr.Error()))
		return &err
	}
	if r.Chrom == "" {
		err := NewValidationError("chrom", "chrom is required")
		return &err
	}
	if !chromRegex.MatchString(r.Chrom) {
		err := NewValidationError("chrom", fmt.Sprintf("invalid chromosome format: %s", r.Chrom))
		return &err
	}
	if r.Start < 0 {
		err := NewValidationError("start", "start must be >= 0")
		return &err
	}
	if r.End <= r.Start {
		err := NewValidationError("end", "end must be greater than start")
		return &err
	}
	if r.PreRenderedWidth < 0 {
		err := NewValidationError("preRenderedWidth", "preRenderedWidth must be >= 0")
		return &err
	}
	return nil
}

type BigBedRequest struct {
	URL   string `json:"url"`
	Chrom string `json:"chrom"`
	Start int    `json:"start"`
	End   int    `json:"end"`
	Type  string `json:"type,omitempty"` // "ccre", "generic" | used for parsing non-universal columns
}

// Validate checks BigBedRequest fields
func (r *BigBedRequest) Validate() *APIError {
	if r.URL == "" {
		err := NewValidationError("url", "url is required")
		return &err
	}
	if _, parseErr := url.ParseRequestURI(r.URL); parseErr != nil {
		err := NewValidationError("url", fmt.Sprintf("invalid url: %s", parseErr.Error()))
		return &err
	}
	if r.Chrom == "" {
		err := NewValidationError("chrom", "chrom is required")
		return &err
	}
	if !chromRegex.MatchString(r.Chrom) {
		err := NewValidationError("chrom", fmt.Sprintf("invalid chromosome format: %s", r.Chrom))
		return &err
	}
	if r.Start < 0 {
		err := NewValidationError("start", "start must be >= 0")
		return &err
	}
	if r.End <= r.Start {
		err := NewValidationError("end", "end must be greater than start")
		return &err
	}
	return nil
}

type TranscriptRequest struct {
	Chrom string `json:"chrom"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Validate checks TranscriptRequest fields
func (r *TranscriptRequest) Validate() *APIError {
	if r.Chrom == "" {
		err := NewValidationError("chrom", "chrom is required")
		return &err
	}
	if !chromRegex.MatchString(r.Chrom) {
		err := NewValidationError("chrom", fmt.Sprintf("invalid chromosome format: %s", r.Chrom))
		return &err
	}
	if r.Start < 0 {
		err := NewValidationError("start", "start must be >= 0")
		return &err
	}
	if r.End <= r.Start {
		err := NewValidationError("end", "end must be greater than start")
		return &err
	}
	return nil
}

// Browser endpoint
type BrowserRequest struct {
	Chrom  string  `json:"chrom"`
	Start  int     `json:"start"`
	End    int     `json:"end"`
	Tracks []Track `json:"tracks"`
}

// Validate checks BrowserRequest fields
func (r *BrowserRequest) Validate() *APIError {
	if r.Chrom == "" {
		err := NewValidationError("chrom", "chrom is required")
		return &err
	}
	if !chromRegex.MatchString(r.Chrom) {
		err := NewValidationError("chrom", fmt.Sprintf("invalid chromosome format: %s", r.Chrom))
		return &err
	}
	if r.Start < 0 {
		err := NewValidationError("start", "start must be >= 0")
		return &err
	}
	if r.End <= r.Start {
		err := NewValidationError("end", "end must be greater than start")
		return &err
	}
	if len(r.Tracks) == 0 {
		err := NewValidationError("tracks", "at least one track is required")
		return &err
	}
	return nil
}

type TrackResponse struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type,omitempty"`
	Data  any    `json:"data"`
	Error string `json:"error,omitempty"`
}

type BrowserResponse struct {
	Data []TrackResponse `json:"data"`
}

// Track configurations
type Track struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

type BigWigConfig struct {
	URL              string `json:"url"`
	PreRenderedWidth int    `json:"preRenderedWidth,omitempty"`
}

type BigBedConfig struct {
	URL  string `json:"url"`
	Type string `json:"type,omitempty"`
}

type Assembly string

const (
	Human Assembly = "grch38"
	Mouse Assembly = "mm10"
)

type TranscriptConfig struct {
	Assembly string `json:"assembly"`
}

func (t *Track) GetBigWigConfig() (BigWigConfig, error) {
	var config BigWigConfig
	err := json.Unmarshal(t.Config, &config)
	return config, err
}

func (t *Track) GetBigBedConfig() (BigBedConfig, error) {
	var config BigBedConfig
	err := json.Unmarshal(t.Config, &config)
	return config, err
}

func (t *Track) GetTranscriptConfig() (TranscriptConfig, error) {
	var config TranscriptConfig
	err := json.Unmarshal(t.Config, &config)
	return config, err
}
