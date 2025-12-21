package main

import (
	"fmt"
	"gb-api/api"
	"net/http"
)

func addRoutes(m *http.ServeMux) {
	m.HandleFunc("/bigwig", api.CORSMiddleware(api.BigWigHandler))
	m.HandleFunc("/bigbed", api.CORSMiddleware(api.BigBedHandler))
	m.HandleFunc("/transcript", api.CORSMiddleware(api.TranscriptHandler))
	m.HandleFunc("/browser", api.CORSMiddleware(api.BrowserHandler))
}

func main() {
	mux := http.NewServeMux()

	addRoutes(mux)
	fmt.Println("Server running on port 8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
