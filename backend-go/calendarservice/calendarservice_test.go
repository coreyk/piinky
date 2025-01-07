package calendarservice

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/coreyk/piinky/backend-go/models"
)

func setupTestConfig(t *testing.T) (string, func()) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "calendar_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a temporary calendar config file
	tempConfig := models.CalendarConfig{
		StartOnSunday: true,
		NumberOfWeeks: 2,
		Calendars: []struct {
			CalendarID string `json:"calendar_id"`
			Color      string `json:"color"`
		}{
			{
				CalendarID: "primary",
				Color:      "1",
			},
		},
	}

	configPath := filepath.Join(tmpDir, "calendar_config.json")
	configFile, err := os.Create(configPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create config file: %v", err)
	}
	defer configFile.Close()

	if err := json.NewEncoder(configFile).Encode(tempConfig); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create a mock credentials file
	credsPath := filepath.Join(tmpDir, "credentials.json")
	credsFile, err := os.Create(credsPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create credentials file: %v", err)
	}
	defer credsFile.Close()

	mockCreds := map[string]interface{}{
		"type":         "service_account",
		"project_id":   "test-project",
		"private_key":  "test-key",
		"client_email": "test@example.com",
	}

	if err := json.NewEncoder(credsFile).Encode(mockCreds); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to write credentials file: %v", err)
	}

	// Set environment variables
	os.Setenv("GOOGLE_CALENDAR_CONFIG_FILE", configPath)
	os.Setenv("GOOGLE_CALENDAR_CREDENTIALS_FILE", credsPath)

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
		os.Unsetenv("GOOGLE_CALENDAR_CONFIG_FILE")
		os.Unsetenv("GOOGLE_CALENDAR_CREDENTIALS_FILE")
	}

	return tmpDir, cleanup
}

func TestHandleGetCalendarMethodNotAllowed(t *testing.T) {
	svc := &Service{}
	req := httptest.NewRequest(http.MethodPost, "/api/calendar", nil)
	w := httptest.NewRecorder()

	svc.HandleGetCalendar(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d for POST request, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestWeekStartCalculation(t *testing.T) {
	// Set up test config
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	testCases := []struct {
		name          string
		startOnSunday bool
		today         time.Time
		expectedStart time.Time
	}{
		{
			name:          "Sunday start on Sunday",
			startOnSunday: true,
			today:         time.Date(2024, 1, 7, 12, 0, 0, 0, time.Local), // A Sunday
			expectedStart: time.Date(2024, 1, 7, 0, 0, 0, 0, time.Local),
		},
		{
			name:          "Sunday start on Wednesday",
			startOnSunday: true,
			today:         time.Date(2024, 1, 10, 12, 0, 0, 0, time.Local), // A Wednesday
			expectedStart: time.Date(2024, 1, 7, 0, 0, 0, 0, time.Local),
		},
		{
			name:          "Monday start on Monday",
			startOnSunday: false,
			today:         time.Date(2024, 1, 8, 12, 0, 0, 0, time.Local), // A Monday
			expectedStart: time.Date(2024, 1, 8, 0, 0, 0, 0, time.Local),
		},
		{
			name:          "Monday start on Thursday",
			startOnSunday: false,
			today:         time.Date(2024, 1, 11, 12, 0, 0, 0, time.Local), // A Thursday
			expectedStart: time.Date(2024, 1, 8, 0, 0, 0, 0, time.Local),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &Service{
				calendarConfig: models.CalendarConfig{
					StartOnSunday: tc.startOnSunday,
				},
				now: func() time.Time {
					return tc.today
				},
			}

			// Create a request with the test time
			req := httptest.NewRequest(http.MethodGet, "/api/calendar", nil)
			w := httptest.NewRecorder()

			// Call the handler
			svc.HandleGetCalendar(w, req)

			// Parse the response
			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Parse the startDate from the response
			startDate, ok := response["startDate"].(string)
			if !ok {
				t.Fatal("startDate not found in response or wrong type")
			}

			parsedStartDate, err := time.Parse(time.RFC3339, startDate)
			if err != nil {
				t.Fatalf("Failed to parse startDate: %v", err)
			}

			// Compare the dates (ignoring time zone differences)
			if !parsedStartDate.Equal(tc.expectedStart) {
				t.Errorf("Expected start date %v, got %v", tc.expectedStart, parsedStartDate)
			}
		})
	}
}
