package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func POST(endpoint string, request any, response any) error {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return err
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	return nil
}

func WriteTOJSON(filepath string, object any) error {
	jsonData, err := json.MarshalIndent(object, "", "	")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ReadFromJSON(filepath string, object any) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &object)
	if err != nil {
		return err
	}

	return nil
}
