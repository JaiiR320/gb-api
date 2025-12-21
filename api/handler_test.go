package api

import (
	"bytes"
	"encoding/json"
	"gb-api/track/bigdata/bigbed"
	"gb-api/track/bigdata/bigwig"
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

	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		t.Error(err.Error())
	}

	var bigWigData []bigwig.BigWigData
	err = json.Unmarshal(dataBytes, &bigWigData)
	if err != nil {
		t.Error(err.Error())
	}

	comp := bigwig.BigWigData{Chr: "chr19", Start: 44905740, End: 44905760, Value: 610.4453}

	if bigWigData[0] != comp {
		t.Errorf("Expected first element to be %+v, got %+v", comp, bigWigData[0])
	}
}

func TestBigBed(t *testing.T) {
	body, err := os.ReadFile("../test/request/bigBedRequest.json")
	if err != nil {
		t.Error(err.Error())
	}

	req := httptest.NewRequest(http.MethodPost, "/bigbed", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	BigBedHandler(w, req)

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

	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		t.Error(err.Error())
	}

	var bigBedData []bigbed.BigBedData
	err = json.Unmarshal(dataBytes, &bigBedData)
	if err != nil {
		t.Error(err.Error())
	}
}
