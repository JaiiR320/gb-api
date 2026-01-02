package api

import (
	"encoding/json"
	"fmt"
	"gb-api/track/bigdata/bigbed"
	"gb-api/track/transcript"
	"log/slog"
	"net/http"
)

func BigWigHandler(w http.ResponseWriter, r *http.Request) {
	uuid := UUID()
	l := slog.With("ID", uuid)
	l.Info("Handling bigwig request")
	TrackHandler(w, r, l, func(req BigWigRequest) (any, error) {
		l.Info("Reading bigwig", "url", req.URL, "chrom", req.Chrom, "start", req.Start, "end", req.End)
		return WigCache.GetCachedWigData(req.URL, req.Chrom, req.Start, req.End)
	})
	l.Info("Finished bigwig request")
}

func BigBedHandler(w http.ResponseWriter, r *http.Request) {
	uuid := UUID()
	l := slog.With("ID", uuid)
	l.Info("Handling bigbed request")
	TrackHandler(w, r, l, func(req BigBedRequest) (any, error) {
		l.Info("Reading bigbed", "url", req.URL, "chrom", req.Chrom, "start", req.Start, "end", req.End)
		data, err := bigbed.ReadBigBed(req.URL, req.Chrom, req.Start, req.End)
		if err != nil {
			return nil, err
		}

		switch req.Type {
		case "ccre":
			return bigbed.ParseCCRE(data)
		default:
			return data, err
		}
	})
	l.Info("Finished bigbed request")
}

func TranscriptHandler(w http.ResponseWriter, r *http.Request) {
	uuid := UUID()
	l := slog.With("ID", uuid)
	l.Info("Handling transcript request")
	TrackHandler(w, r, l, func(req TranscriptRequest) (any, error) {
		l.Info("Getting transcripts", "chrom", req.Chrom, "start", req.Start, "end", req.End)
		return transcript.GetTranscripts(req.Chrom, req.Start, req.End)
	})
	l.Info("Finished transcript request")
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

	// Track data fetchers
	switch t.Type {
	case "bigwig":
		cfg, err := t.GetBigWigConfig()
		if err != nil {
			err = fmt.Errorf("Could not get BigWig config, %w", err)
			break
		}
		logger.Info("Reading bigWig", "url", cfg.URL, "chrom", request.Chrom, "start", request.Start, "end", request.End)
		data, err = WigCache.GetCachedWigData(cfg.URL, request.Chrom, request.Start, request.End)
	case "bigbed":
		cfg, err := t.GetBigBedConfig()
		if err != nil {
			err = fmt.Errorf("Could not get BigBedconfig, %w", err)
			break
		}
		logger.Info("Reading bigBed", "url", cfg.URL, "chrom", request.Chrom, "start", request.Start, "end", request.End)
		data, err = bigbed.ReadBigBed(cfg.URL, request.Chrom, request.Start, request.End)
	case "transcript":
		_, err := t.GetTranscriptConfig()
		if err != nil {
			err = fmt.Errorf("Could not get Transcript config, %w", err)
			break
		}
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

	results <- TrackResponse{
		ID:   t.ID,
		Type: t.Type,
		Data: data,
	}
}
