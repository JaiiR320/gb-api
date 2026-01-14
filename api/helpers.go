package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
)

func UUID() string {
	src := make([]byte, 8)
	n, _ := rand.Read(src) // ignore error as per docs

	dst := make([]byte, hex.EncodedLen(n))
	hex.Encode(dst, src)

	str := string(dst)

	return str[0:4] + "-" + str[4:8] + "-" + str[8:12] + "-" + str[12:]
}

// WriteJSONError writes a standardized JSON error response
func WriteJSONError(w http.ResponseWriter, requestID string, statusCode int, apiErr APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: apiErr})
}

// Wrapper function for making new track-specific handlers
func TrackHandler[Req Validatable, Data any](w http.ResponseWriter, r *http.Request, l *slog.Logger, requestID string, fetch func(req Req) (Data, error)) {
	if r.Method != http.MethodPost {
		WriteJSONError(w, requestID, http.StatusMethodNotAllowed,
			NewAPIError(ErrCodeMethodNotAllowed, "Method not allowed"))
		l.Error("Method not allowed", "method", r.Method)
		return
	}

	var request Req
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteJSONError(w, requestID, http.StatusBadRequest,
			APIError{Code: ErrCodeInvalidJSON, Message: "Failed to decode request", Details: err.Error()})
		l.Error("Failed to decode request", "error", err)
		return
	}

	// Validate the request
	if validationErr := request.Validate(); validationErr != nil {
		WriteJSONError(w, requestID, http.StatusBadRequest, *validationErr)
		l.Error("Validation failed", "field", validationErr.Field, "message", validationErr.Message)
		return
	}

	data, err := fetch(request)
	if err != nil {
		WriteJSONError(w, requestID, http.StatusInternalServerError,
			APIError{Code: ErrCodeInternalError, Message: "Failed to fetch data", Details: err.Error()})
		l.Error("Failed to fetch data", "error", err)
		return
	}

	response := TrackResponse{
		Data: data,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		WriteJSONError(w, requestID, http.StatusInternalServerError,
			NewAPIError(ErrCodeInternalError, "Failed to encode response"))
		l.Error("Failed to encode response", "error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	if _, err = w.Write(responseBytes); err != nil {
		l.Error("Failed to write response", "error", err)
	}
}
