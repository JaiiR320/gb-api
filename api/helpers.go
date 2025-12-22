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

// Wrapper function for making new track-specific handlers
func TrackHandler[Req any, Data any](w http.ResponseWriter, r *http.Request, l *slog.Logger, fetch func(req Req) (Data, error)) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		l.Error("Method not allowed", "method", r.Method)
		return
	}

	var request Req
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		l.Error("Failed to decode request", "error", err)
		return
	}

	data, err := fetch(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		l.Error("Failed to data", "error", err)
		return
	}

	response := TrackResponse{
		Data: data,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		l.Error("Failed to encode response", "error", err)
		return
	}
	w.Header().Set("X-Cache-Status", "HIT") // or "MISS"
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(responseBytes); err != nil {
		l.Error("Failed to write response", "error", err)
	}
}
