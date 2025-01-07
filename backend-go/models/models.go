package models

type CalendarConfig struct {
	Calendars []struct {
		CalendarID string `json:"calendar_id"`
		Color      string `json:"color"`
	} `json:"calendars"`
	NumberOfWeeks int  `json:"number_of_weeks"`
	StartOnSunday bool `json:"start_on_sunday"`
}

type WeatherConfig struct {
	Weather struct {
		APIKey   string  `json:"api_key"`
		Lat      float64 `json:"lat"`
		Lon      float64 `json:"lon"`
		Location string  `json:"location"`
		Units    string  `json:"units"`
	} `json:"weather"`
}

type WeatherData struct {
	Latitude       float64         `json:"lat"`
	Longitude      float64         `json:"lon"`
	Temperature    TemperatureData `json:"temperature"`
	Status         string          `json:"status"`
	DetailedStatus string          `json:"detailed_status"`
	Icon           int             `json:"icon"`
	Humidity       float64         `json:"humidity"`
	WindSpeed      float64         `json:"wind_speed"`
	WindDir        float64         `json:"wind_dir"`
	UVI            float64         `json:"uvi"`
	Clouds         float64         `json:"clouds"`
	Summary        string          `json:"summary"`
	Location       string          `json:"location"`
	HourlyForecast []ForecastData  `json:"hourly_forecast"`
	DailyForecast  []ForecastData  `json:"daily_forecast"`
}

type ForecastData struct {
	Timestamp   int64           `json:"timestamp"`
	Temperature TemperatureData `json:"temperature"`
	Status      string          `json:"status"`
	Icon        int             `json:"icon"`
	Humidity    int             `json:"humidity"`
	WindSpeed   float64         `json:"wind_speed"`
}

type TemperatureData struct {
	Temp      float64 `json:"temp"`
	TempMin   float64 `json:"min"`
	TempMax   float64 `json:"max"`
	FeelsLike float64 `json:"feels_like"`
}

type OWMWeatherData struct {
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	Timezone       string  `json:"timezone"`
	TimezoneOffset int64   `json:"timezone_offset"`
	Current        struct {
		Timestamp  int64          `json:"dt"`
		Sunrise    int64          `json:"sunrise"`
		Sunset     int64          `json:"sunset"`
		Temp       float64        `json:"temp"`
		FeelsLike  float64        `json:"feels_like"`
		Humidity   int            `json:"humidity"`
		DewPoint   float64        `json:"dew_point"`
		UVI        float64        `json:"uvi"`
		Clouds     float64        `json:"clouds"`
		Visibility float64        `json:"visibility"`
		WindSpeed  float64        `json:"wind_speed"`
		WindDeg    float64        `json:"wind_deg"`
		WindGust   float64        `json:"wind_gust"`
		Weather    []OWMCondition `json:"weather"`
	} `json:"current"`
	Hourly []OWMHourlyForecast `json:"hourly"`
	Daily  []OWMDailyForecast  `json:"daily"`
}

type OWMHourlyForecast struct {
	Timestamp  int64          `json:"dt"`
	Temp       float64        `json:"temp"`
	FeelsLike  float64        `json:"feels_like"`
	Pressure   float64        `json:"pressure"`
	Humidity   int            `json:"humidity"`
	DewPoint   float64        `json:"dew_point"`
	UVI        float64        `json:"uvi"`
	Clouds     float64        `json:"clouds"`
	Visibility float64        `json:"visibility"`
	WindSpeed  float64        `json:"wind_speed"`
	WindDeg    float64        `json:"wind_deg"`
	WindGust   float64        `json:"wind_gust"`
	Weather    []OWMCondition `json:"weather"`
	Pop        float64        `json:"pop"`
}

type OWMDailyForecast struct {
	Timestamp int64   `json:"dt"`
	Sunrise   int64   `json:"sunrise"`
	Sunset    int64   `json:"sunset"`
	Moonrise  int64   `json:"moonrise"`
	Moonset   int64   `json:"moonset"`
	MoonPhase float64 `json:"moon_phase"`
	Summary   string  `json:"summary"`
	Temp      struct {
		Day   float64 `json:"day"`
		Min   float64 `json:"min"`
		Max   float64 `json:"max"`
		Night float64 `json:"night"`
		Eve   float64 `json:"eve"`
		Morn  float64 `json:"morn"`
	} `json:"temp"`
	FeelsLike struct {
		Day   float64 `json:"day"`
		Night float64 `json:"night"`
		Eve   float64 `json:"eve"`
		Morn  float64 `json:"morn"`
	} `json:"feels_like"`
	Pressure  float64        `json:"pressure"`
	Humidity  int            `json:"humidity"`
	DewPoint  float64        `json:"dew_point"`
	WindSpeed float64        `json:"wind_speed"`
	WindDeg   float64        `json:"wind_deg"`
	WindGust  float64        `json:"wind_gust"`
	Weather   []OWMCondition `json:"weather"`
	Clouds    float64        `json:"clouds"`
	Pop       float64        `json:"pop"`
	UVI       float64        `json:"uvi"`
}

type OWMCondition struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}
