export interface WeatherData {
  temperature: {
    temp: number;
    feels_like: number;
    min: number;
    max: number;
  };
  status: string;
  icon: string;
  humidity: number;
  wind_speed: number;
  summary: string;
  hourly_forecast: ForecastData[];
  daily_forecast: ForecastData[];
}

export interface ForecastData {
  timestamp: string;
  temperature: {
    temp: number;
    feels_like: number;
    min: number;
    max: number;
  };
  status: string;
  icon: string;
  humidity: number;
  wind_speed: number;
}
