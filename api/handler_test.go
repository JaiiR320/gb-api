package api

import (
	"bytes"
	"encoding/json"
	"gb-api/track/bigwig"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBigWigHandler(t *testing.T) {
	body, err := os.ReadFile("../test/request/bigWigRequest.json")
	if err != nil {
		t.Error(err.Error())
	}

	req := httptest.NewRequest(http.MethodPost, "/bigwig", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	BigWigHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err.Error())
	}

	var response TrackResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		t.Error(err.Error())
	}

	var bigWigData []bigwig.BigWigData
	err = json.Unmarshal(response.Data, &bigWigData)
	if err != nil {
		t.Error(err.Error())
	}

	if len(bigWigData) != 101 {
		t.Fail()
	}
}
