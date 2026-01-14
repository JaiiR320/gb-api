package api

import (
	"encoding/json"
	"fmt"
	"gb-api/track/bigdata/bigbed"
	"gb-api/track/bigdata/bigwig"
	"gb-api/track/transcript"
	"log/slog"
	"net/http"
)

func BigWigHandler(w http.ResponseWriter, r *http.Request) {
	uuid := UUID()
	l := slog.With("ID", uuid)
	l.Info("Handling bigwig request")
	TrackHandler(w, r, l, uuid, func(req BigWigRequest) (any, error) {
		l.Info("Reading bigwig", "url", req.URL, "chrom", req.Chrom, "start", req.Start, "end", req.End, "preRenderedWidth", req.PreRenderedWidth)
		data, err := bigwig.GetCachedWigData(req.URL, req.Chrom, req.Start, req.End, req.PreRenderedWidth)
		if err != nil {
			return nil, err
		}

		// Resample to prerendered width if specified
		if req.PreRenderedWidth > 0 {
			return bigwig.ResampleToWidth(data, req.PreRenderedWidth), nil
		}
		return data, nil
	})
	l.Info("Finished bigwig request")
}

func BigBedHandler(w http.ResponseWriter, r *http.Request) {
	uuid := UUID()
	l := slog.With("ID", uuid)
	l.Info("Handling bigbed request")
	TrackHandler(w, r, l, uuid, func(req BigBedRequest) (any, error) {
		l.Info("Reading bigbed", "url", req.URL, "chrom", req.Chrom, "start", req.Start, "end", req.End)
		data, err := bigbed.GetCachedBedData(req.URL, req.Chrom, req.Start, req.End)
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
	TrackHandler(w, r, l, uuid, func(req TranscriptRequest) (any, error) {
		l.Info("Getting transcripts", "chrom", req.Chrom, "start", req.Start, "end", req.End)
		data, err := transcript.GetTranscripts(req.Chrom, req.Start, req.End)

		const defaultPaddingBp = 100
		return transcript.LegacyWithLayout(data, defaultPaddingBp, err)
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
	w.Header().Set("X-Request-ID", uuid)
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
		logger.Info("Reading bigWig", "url", cfg.URL, "chrom", request.Chrom, "start", request.Start, "end", request.End, "preRenderedWidth", cfg.PreRenderedWidth)
		wigData, err := bigwig.GetCachedWigData(cfg.URL, request.Chrom, request.Start, request.End, cfg.PreRenderedWidth)
		if err != nil {
			break
		}

		// Resample to prerendered width if specified
		if cfg.PreRenderedWidth > 0 {
			data = bigwig.ResampleToWidth(wigData, cfg.PreRenderedWidth)
		} else {
			data = wigData
		}
	case "bigbed":
		cfg, err := t.GetBigBedConfig()
		if err != nil {
			err = fmt.Errorf("Could not get BigBedconfig, %w", err)
			break
		}
		logger.Info("Reading bigBed", "url", cfg.URL, "chrom", request.Chrom, "start", request.Start, "end", request.End)
		data, err = bigbed.GetCachedBedData(cfg.URL, request.Chrom, request.Start, request.End)
	case "transcript":
		_, err := t.GetTranscriptConfig()
		if err != nil {
			err = fmt.Errorf("Could not get Transcript config, %w", err)
			break
		}
		logger.Info("Getting transcripts", "chrom", request.Chrom, "start", request.Start, "end", request.End)
		genes, err := transcript.GetTranscripts(request.Chrom, request.Start, request.End)
		if err != nil {
			break
		}

		// Use default padding of 100bp for layout
		const defaultPaddingBp = 100
		data, err = transcript.LegacyWithLayout(genes, defaultPaddingBp, nil)
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
