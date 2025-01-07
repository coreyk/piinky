package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServer(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create new server: %v", err)
	}

	if server.mux == nil {
		t.Error("Server mux is nil")
	}
	if server.calendarSvc == nil {
		t.Error("Calendar service is nil")
	}
	if server.weatherSvc == nil {
		t.Error("Weather service is nil")
	}
}

func TestCORSMiddleware(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create new server: %v", err)
	}

	// Create a test handler that we'll wrap with CORS
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a test server with the CORS-wrapped handler
	ts := httptest.NewServer(server.cors(testHandler))
	defer ts.Close()

	// Test OPTIONS request
	req, err := http.NewRequest(http.MethodOptions, ts.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check CORS headers
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type",
	}

	for header, expected := range expectedHeaders {
		if got := resp.Header.Get(header); got != expected {
			t.Errorf("Expected header %s to be %s, got %s", header, expected, got)
		}
	}

	// Check status code for OPTIONS request
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestSetupRoutes(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("Failed to create new server: %v", err)
	}

	server.setupRoutes()

	// Test that routes are set up by making test requests
	testCases := []struct {
		name     string
		path     string
		method   string
		wantCode int
	}{
		{
			name:     "Calendar API",
			path:     "/api/calendar",
			method:   http.MethodGet,
			wantCode: http.StatusOK,
		},
		{
			name:     "Weather API",
			path:     "/api/weather",
			method:   http.MethodGet,
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			server.mux.ServeHTTP(w, req)

			// We're just checking if the routes are registered
			// The actual response code might be different in real execution
			if w.Code == http.StatusNotFound {
				t.Errorf("Route %s not found", tc.path)
			}
		})
	}
}
