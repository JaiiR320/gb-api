package test

import (
	"gb-api/api"
	"testing"
)

func TestBigWigHandler(t *testing.T) {
	var request api.BigWigRequest

	err := ReadFromJSON("./request/bigWigRequest.json", &request)
	if err != nil {
		t.Error(err.Error())
	}

	var response api.TrackResponse
	err = POST("http://localhost:8080/bigwig", request, &response)
	if err != nil {
		t.Error(err.Error())
	}

	err = WriteTOJSON("./response/bigwigResponse.json", &response)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestTranscriptHandler(t *testing.T) {
	var request api.TranscriptRequest
	err := ReadFromJSON("./request/transcriptRequest.json", &request)
	if err != nil {
		t.Error(err.Error())
	}

	var response api.TrackResponse
	err = POST("http://localhost:8080/transcript", request, &response)
	if err != nil {
		t.Error(err.Error())
	}

	err = WriteTOJSON("./response/transcriptResponse.json", response)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestBrowserHandler(t *testing.T) {
	var request api.BrowserRequest
	err := ReadFromJSON("./request/browserRequest.json", &request)
	if err != nil {
		t.Error(err.Error())
	}

	var response api.BrowserResponse
	err = POST("http://localhost:8080/browser", request, &response)
	if err != nil {
		t.Error(err.Error())
	}

	err = WriteTOJSON("./response/browserResponse.json", response)
	if err != nil {
		t.Error(err.Error())
	}
}
