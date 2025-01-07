package weatherservice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/coreyk/piinky/backend-go/models"
)

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Service struct {
	weatherConfig models.WeatherConfig
	httpClient    HTTPClient
}

func NewService() (*Service, error) {
	weatherConfigFile := os.Getenv("OWM_CONFIG_FILE")
	if weatherConfigFile == "" {
		weatherConfigFile = "../owm_config.json"
		fmt.Println("OWM_CONFIG_FILE is not set in the .env file, using default: ", weatherConfigFile)
	}

	weatherConfigBytes, err := os.ReadFile(weatherConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read weather API key: %v", err)
	}

	var weatherConfig models.WeatherConfig
	if err := json.Unmarshal(weatherConfigBytes, &weatherConfig); err != nil {
		return nil, fmt.Errorf("failed to parse weather API key: %v", err)
	}

	return &Service{
		weatherConfig: weatherConfig,
		httpClient:    http.DefaultClient,
	}, nil
}

// SetHTTPClient allows setting a custom HTTP client (useful for testing)
func (s *Service) SetHTTPClient(client HTTPClient) {
	s.httpClient = client
}

func (s *Service) HandleGetWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	owmData, err := s.fetchWeatherData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	weatherResponse := s.mapToWeatherData(owmData)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(weatherResponse); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Service) fetchWeatherData() (*models.OWMWeatherData, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&exclude=minutely,alerts&units=%s&appid=%s",
		s.weatherConfig.Weather.Lat,
		s.weatherConfig.Weather.Lon,
		s.weatherConfig.Weather.Units,
		s.weatherConfig.Weather.APIKey,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current weather: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch current weather: status code %d", resp.StatusCode)
	}

	var owmData models.OWMWeatherData
	if err := json.NewDecoder(resp.Body).Decode(&owmData); err != nil {
		return nil, fmt.Errorf("failed to parse weather data: %v", err)
	}

	return &owmData, nil
}

func (s *Service) mapToWeatherData(owmData *models.OWMWeatherData) models.WeatherData {
	weatherData := models.WeatherData{
		Latitude:  owmData.Lat,
		Longitude: owmData.Lon,
		Temperature: models.TemperatureData{
			Temp:      owmData.Current.Temp,
			TempMin:   owmData.Daily[0].Temp.Min,
			TempMax:   owmData.Daily[0].Temp.Max,
			FeelsLike: owmData.Current.FeelsLike,
		},
		Status:         owmData.Current.Weather[0].Main,
		DetailedStatus: owmData.Current.Weather[0].Description,
		Icon:           owmData.Current.Weather[0].ID,
		Humidity:       float64(owmData.Current.Humidity),
		WindSpeed:      owmData.Current.WindSpeed,
		WindDir:        owmData.Current.WindDeg,
		UVI:            owmData.Current.UVI,
		Clouds:         owmData.Current.Clouds,
		Summary:        owmData.Current.Weather[0].Main,
		Location:       s.weatherConfig.Weather.Location,
	}

	weatherData.HourlyForecast = mapHourlyForecast(owmData.Hourly)
	weatherData.DailyForecast = mapDailyForecast(owmData.Daily)

	return weatherData
}

func mapHourlyForecast(hourlyData []models.OWMHourlyForecast) []models.ForecastData {
	forecast := make([]models.ForecastData, len(hourlyData))
	for i, hourly := range hourlyData {
		forecast[i] = models.ForecastData{
			Timestamp: hourly.Timestamp,
			Temperature: models.TemperatureData{
				Temp:      hourly.Temp,
				FeelsLike: hourly.FeelsLike,
			},
			Status:    hourly.Weather[0].Main,
			Icon:      hourly.Weather[0].ID,
			Humidity:  hourly.Humidity,
			WindSpeed: hourly.WindSpeed,
		}
	}
	return forecast
}

func mapDailyForecast(dailyData []models.OWMDailyForecast) []models.ForecastData {
	forecast := make([]models.ForecastData, len(dailyData))
	for i, daily := range dailyData {
		forecast[i] = models.ForecastData{
			Timestamp: daily.Timestamp,
			Temperature: models.TemperatureData{
				Temp:      daily.Temp.Day,
				TempMin:   daily.Temp.Min,
				TempMax:   daily.Temp.Max,
				FeelsLike: daily.FeelsLike.Day,
			},
			Status: daily.Weather[0].Main,
			Icon:   daily.Weather[0].ID,
		}
	}
	return forecast
}
