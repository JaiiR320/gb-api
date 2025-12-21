package api

import (
	"encoding/json"
	"fmt"
	"gb-api/track/bigwig"
	"gb-api/track/transcript"
	"log/slog"
	"net/http"
)

// BigWigHandler accepts arbitrary BigWig URLs to enable users to visualize
// genomic data from any source: public databases, company servers, or local files.
// This flexibility is intentional to support experimentation and personal data exploration.
// Users can provide URLs pointing to:
//   - Public databases (UCSC, ENCODE, etc.)
//   - Company internal servers (http://internal-server/data.bigwig)
//   - Local files (http://localhost:8000/mydata.bigwig)
//
// URLs are validated to be BigWig format via header checking, but no domain
// restrictions are applied to preserve maximum flexibility.
func BigWigHandler(w http.ResponseWriter, r *http.Request) {
	uuid := UUID()
	logger := slog.With("ID", uuid)
	logger.Info("Handling bigwig request")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Error("Method not allowed", "method", r.Method)
		return
	}

	var request BigWigRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		logger.Error("Failed to decode request", "error", err)
		return
	}

	logger.Info("Reading bigwig", "url", request.URL, "chrom", request.Chrom, "start", request.Start, "end", request.End)
	data, err := bigwig.ReadBigWig(request.URL, request.Chrom, request.Start, request.End)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("Failed to get bigWig data", "error", err, "url", request.URL, "chrom", request.Chrom, "start", request.Start, "end", request.End)
		return
	}

	d, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to marshal data", http.StatusInternalServerError)
		logger.Error("Failed to marshal data", "error", err)
		return
	}

	response := TrackResponse{
		Data: d,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		logger.Error("Failed to encode response", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(responseBytes); err != nil {
		logger.Error("Failed to write response", "error", err)
	}

	logger.Info("Finished bigwig request")
}

func TranscriptHandler(w http.ResponseWriter, r *http.Request) {
	uuid := UUID()
	logger := slog.With("ID", uuid)
	logger.Info("Handling transcript request")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Error("Method not allowed", "method", r.Method)
		return
	}

	var request TranscriptRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		logger.Error("Failed to decode request", "error", err)
		return
	}

	logger.Info("Getting transcripts", "chrom", request.Chrom, "start", request.Start, "end", request.End)
	data, err := transcript.GetTranscripts(request.Chrom, request.Start, request.End)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("Failed to get transcript data", "error", err, "chrom", request.Chrom, "start", request.Start, "end", request.End)
		return
	}

	d, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to marshal data", http.StatusInternalServerError)
		logger.Error("Failed to marshal data", "error", err)
		return
	}

	response := TrackResponse{
		Data: d,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		logger.Error("Failed to encode response", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(responseBytes); err != nil {
		logger.Error("Failed to write response", "error", err)
	}

	logger.Info("Finished transcript request")
}

func BrowserHandler(w http.ResponseWriter, r *http.Request) {
	uuid := UUID()
	logger := slog.With("ID", uuid)
	logger.Info("Handling browser request")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Error("Method not allowed", "method", r.Method)
		return
	}

	var request BrowserRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Failed to decode request", http.StatusBadRequest)
		logger.Error("Failed to decode request", "error", err)
		return
	}

	var results = make(chan TrackResponse, len(request.Tracks))

	for _, track := range request.Tracks {
		go getTrackData(track, request, results)
	}

	var responses []TrackResponse
	for i := 0; i < len(request.Tracks); i++ {
		responses = append(responses, <-results)
	}

	response := BrowserResponse{
		Data: responses,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		logger.Error("Failed to encode response", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(responseBytes); err != nil {
		logger.Error("Failed to write response", "error", err)
	}
	logger.Info("Finished browser request")
}

func getTrackData(t Track, request BrowserRequest, results chan TrackResponse) {
	logger := slog.With("track", t.ID)

	var data any
	var err error

	switch t.Type {
	case "bigwig":
		cfg, err := t.GetBigWigConfig()
		if err != nil {
			err = fmt.Errorf("Could not get BigWig config, %w", err)
			break
		}
		logger.Info("Reading bigWig", "url", cfg.URL, "chrom", request.Chrom, "start", request.Start, "end", request.End)
		data, err = bigwig.ReadBigWig(cfg.URL, request.Chrom, request.Start, request.End)
	case "transcript":
		logger.Info("Getting transcripts", "chrom", request.Chrom, "start", request.Start, "end", request.End)
		data, err = transcript.GetTranscripts(request.Chrom, request.Start, request.End)
	default:
		err = fmt.Errorf("Invalid track type %s", t.Type)
	}

	if err != nil {
		results <- TrackResponse{
			ID:    t.ID,
			Type:  t.Type,
			Error: err.Error(),
		}
		logger.Error("Error getting data", "error", err)
		return
	}

	d, err := json.Marshal(data)
	if err != nil {
		results <- TrackResponse{
			ID:    t.ID,
			Type:  t.Type,
			Error: fmt.Sprintf("Failed to marshal data: %v", err),
		}
		logger.Error("Failed to marshal data", "error", err)
		return
	}

	results <- TrackResponse{
		ID:   t.ID,
		Type: t.Type,
		Data: d,
	}

}
