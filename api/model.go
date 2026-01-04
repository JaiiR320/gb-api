package api

import (
	"encoding/json"
)

// Track requests
type BigWigRequest struct {
	URL              string `json:"url"`
	Chrom            string `json:"chrom"`
	Start            int    `json:"start"`
	End              int    `json:"end"`
	PreRenderedWidth int    `json:"preRenderedWidth,omitempty"` // Number of points to return
}

type BigBedRequest struct {
	URL   string `json:"url"`
	Chrom string `json:"chrom"`
	Start int    `json:"start"`
	End   int    `json:"end"`
	Type  string `json:"type"` // "ccre", "generic"
}

type TranscriptRequest struct {
	Chrom string `json:"chrom"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Browser endpoint
type BrowserRequest struct {
	Chrom  string  `json:"chrom"`
	Start  int     `json:"start"`
	End    int     `json:"end"`
	Tracks []Track `json:"tracks"`
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
