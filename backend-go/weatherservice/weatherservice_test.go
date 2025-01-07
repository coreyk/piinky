package weatherservice

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/coreyk/piinky/backend-go/models"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func setupTestConfig(t *testing.T) (string, func()) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "weather_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a temporary weather config file
	tempConfig := models.WeatherConfig{
		Weather: struct {
			APIKey   string  `json:"api_key"`
			Lat      float64 `json:"lat"`
			Lon      float64 `json:"lon"`
			Location string  `json:"location"`
			Units    string  `json:"units"`
		}{
			APIKey:   "test-api-key",
			Lat:      37.7749,
			Lon:      -122.4194,
			Location: "San Francisco",
			Units:    "metric",
		},
	}

	configPath := filepath.Join(tmpDir, "weather_config.json")
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

	// Set environment variables
	os.Setenv("OWM_CONFIG_FILE", configPath)

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
		os.Unsetenv("OWM_CONFIG_FILE")
	}

	return tmpDir, cleanup
}

func TestHandleGetWeatherMethodNotAllowed(t *testing.T) {
	svc := &Service{}
	req := httptest.NewRequest(http.MethodPost, "/api/weather", nil)
	w := httptest.NewRecorder()

	svc.HandleGetWeather(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d for POST request, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestMapToWeatherData(t *testing.T) {
	owmData := &models.OWMWeatherData{
		Lat: 37.7749,
		Lon: -122.4194,
		Current: struct {
			Timestamp  int64                 `json:"dt"`
			Sunrise    int64                 `json:"sunrise"`
			Sunset     int64                 `json:"sunset"`
			Temp       float64               `json:"temp"`
			FeelsLike  float64               `json:"feels_like"`
			Humidity   int                   `json:"humidity"`
			DewPoint   float64               `json:"dew_point"`
			UVI        float64               `json:"uvi"`
			Clouds     float64               `json:"clouds"`
			Visibility float64               `json:"visibility"`
			WindSpeed  float64               `json:"wind_speed"`
			WindDeg    float64               `json:"wind_deg"`
			WindGust   float64               `json:"wind_gust"`
			Weather    []models.OWMCondition `json:"weather"`
		}{
			Timestamp: 1641456000,
			Temp:      20.5,
			FeelsLike: 19.8,
			Humidity:  65,
			UVI:       4.5,
			Clouds:    75,
			WindSpeed: 3.5,
			WindDeg:   180,
			Weather: []models.OWMCondition{
				{
					ID:          800,
					Main:        "Clear",
					Description: "clear sky",
					Icon:        "01d",
				},
			},
		},
		Daily: []models.OWMDailyForecast{
			{
				Timestamp: 1641456000,
				Temp: struct {
					Day   float64 `json:"day"`
					Min   float64 `json:"min"`
					Max   float64 `json:"max"`
					Night float64 `json:"night"`
					Eve   float64 `json:"eve"`
					Morn  float64 `json:"morn"`
				}{
					Day: 20.5,
					Min: 15.0,
					Max: 25.0,
				},
				Weather: []models.OWMCondition{
					{
						ID:   800,
						Main: "Clear",
					},
				},
			},
		},
		Hourly: []models.OWMHourlyForecast{
			{
				Timestamp: 1641456000,
				Temp:      20.5,
				FeelsLike: 19.8,
				Humidity:  65,
				WindSpeed: 3.5,
				Weather: []models.OWMCondition{
					{
						ID:   800,
						Main: "Clear",
					},
				},
			},
		},
	}

	svc := &Service{}
	weatherData := svc.mapToWeatherData(owmData)

	// Test mapped data
	if weatherData.Latitude != owmData.Lat {
		t.Errorf("Expected latitude %f, got %f", owmData.Lat, weatherData.Latitude)
	}
	if weatherData.Longitude != owmData.Lon {
		t.Errorf("Expected longitude %f, got %f", owmData.Lon, weatherData.Longitude)
	}
	if weatherData.Temperature.Temp != owmData.Current.Temp {
		t.Errorf("Expected temperature %f, got %f", owmData.Current.Temp, weatherData.Temperature.Temp)
	}
	if weatherData.Status != owmData.Current.Weather[0].Main {
		t.Errorf("Expected status %s, got %s", owmData.Current.Weather[0].Main, weatherData.Status)
	}
}

func TestMapHourlyForecast(t *testing.T) {
	hourlyData := []models.OWMHourlyForecast{
		{
			Timestamp: 1641456000,
			Temp:      20.5,
			FeelsLike: 19.8,
			Humidity:  65,
			WindSpeed: 3.5,
			Weather: []models.OWMCondition{
				{
					ID:   800,
					Main: "Clear",
				},
			},
		},
	}

	forecast := mapHourlyForecast(hourlyData)

	if len(forecast) != len(hourlyData) {
		t.Errorf("Expected %d forecasts, got %d", len(hourlyData), len(forecast))
	}

	if forecast[0].Timestamp != hourlyData[0].Timestamp {
		t.Errorf("Expected timestamp %d, got %d", hourlyData[0].Timestamp, forecast[0].Timestamp)
	}
	if forecast[0].Temperature.Temp != hourlyData[0].Temp {
		t.Errorf("Expected temperature %f, got %f", hourlyData[0].Temp, forecast[0].Temperature.Temp)
	}
	if forecast[0].Status != hourlyData[0].Weather[0].Main {
		t.Errorf("Expected status %s, got %s", hourlyData[0].Weather[0].Main, forecast[0].Status)
	}
}

func TestMapDailyForecast(t *testing.T) {
	dailyData := []models.OWMDailyForecast{
		{
			Timestamp: 1641456000,
			Temp: struct {
				Day   float64 `json:"day"`
				Min   float64 `json:"min"`
				Max   float64 `json:"max"`
				Night float64 `json:"night"`
				Eve   float64 `json:"eve"`
				Morn  float64 `json:"morn"`
			}{
				Day: 20.5,
				Min: 15.0,
				Max: 25.0,
			},
			FeelsLike: struct {
				Day   float64 `json:"day"`
				Night float64 `json:"night"`
				Eve   float64 `json:"eve"`
				Morn  float64 `json:"morn"`
			}{
				Day: 19.8,
			},
			Weather: []models.OWMCondition{
				{
					ID:   800,
					Main: "Clear",
				},
			},
		},
	}

	forecast := mapDailyForecast(dailyData)

	if len(forecast) != len(dailyData) {
		t.Errorf("Expected %d forecasts, got %d", len(dailyData), len(forecast))
	}

	if forecast[0].Timestamp != dailyData[0].Timestamp {
		t.Errorf("Expected timestamp %d, got %d", dailyData[0].Timestamp, forecast[0].Timestamp)
	}
	if forecast[0].Temperature.Temp != dailyData[0].Temp.Day {
		t.Errorf("Expected temperature %f, got %f", dailyData[0].Temp.Day, forecast[0].Temperature.Temp)
	}
	if forecast[0].Temperature.TempMin != dailyData[0].Temp.Min {
		t.Errorf("Expected min temperature %f, got %f", dailyData[0].Temp.Min, forecast[0].Temperature.TempMin)
	}
	if forecast[0].Temperature.TempMax != dailyData[0].Temp.Max {
		t.Errorf("Expected max temperature %f, got %f", dailyData[0].Temp.Max, forecast[0].Temperature.TempMax)
	}
	if forecast[0].Status != dailyData[0].Weather[0].Main {
		t.Errorf("Expected status %s, got %s", dailyData[0].Weather[0].Main, forecast[0].Status)
	}
}

func TestHandleGetWeather(t *testing.T) {
	// Set up test config
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Create mock response data
	mockResponse := models.OWMWeatherData{
		Lat: 37.7749,
		Lon: -122.4194,
		Current: struct {
			Timestamp  int64                 `json:"dt"`
			Sunrise    int64                 `json:"sunrise"`
			Sunset     int64                 `json:"sunset"`
			Temp       float64               `json:"temp"`
			FeelsLike  float64               `json:"feels_like"`
			Humidity   int                   `json:"humidity"`
			DewPoint   float64               `json:"dew_point"`
			UVI        float64               `json:"uvi"`
			Clouds     float64               `json:"clouds"`
			Visibility float64               `json:"visibility"`
			WindSpeed  float64               `json:"wind_speed"`
			WindDeg    float64               `json:"wind_deg"`
			WindGust   float64               `json:"wind_gust"`
			Weather    []models.OWMCondition `json:"weather"`
		}{
			Timestamp: 1641456000,
			Temp:      20.5,
			FeelsLike: 19.8,
			Humidity:  65,
			Weather: []models.OWMCondition{
				{
					ID:          800,
					Main:        "Clear",
					Description: "clear sky",
				},
			},
		},
		Daily: []models.OWMDailyForecast{
			{
				Timestamp: 1641456000,
				Temp: struct {
					Day   float64 `json:"day"`
					Min   float64 `json:"min"`
					Max   float64 `json:"max"`
					Night float64 `json:"night"`
					Eve   float64 `json:"eve"`
					Morn  float64 `json:"morn"`
				}{
					Day: 20.5,
					Min: 15.0,
					Max: 25.0,
				},
				Weather: []models.OWMCondition{
					{
						ID:   800,
						Main: "Clear",
					},
				},
			},
		},
		Hourly: []models.OWMHourlyForecast{
			{
				Timestamp: 1641456000,
				Temp:      20.5,
				FeelsLike: 19.8,
				Humidity:  65,
				Weather: []models.OWMCondition{
					{
						ID:   800,
						Main: "Clear",
					},
				},
			},
		},
	}

	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			responseBody, err := json.Marshal(mockResponse)
			if err != nil {
				t.Fatalf("Failed to marshal mock response: %v", err)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
				Header:     make(http.Header),
			}, nil
		},
	}

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/api/weather", nil)
	w := httptest.NewRecorder()

	// Create service and handle request
	svc, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Set the mock client
	svc.SetHTTPClient(mockClient)

	svc.HandleGetWeather(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var response models.WeatherData
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify basic response structure
	if response.Status == "" {
		t.Error("Expected non-empty status")
	}
	if response.Temperature.Temp == 0 {
		t.Error("Expected non-zero temperature")
	}
}
