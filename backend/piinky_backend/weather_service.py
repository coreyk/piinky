from dataclasses import dataclass
from typing import Dict, List, Optional
import json
import httpx
from fastapi import HTTPException

@dataclass
class WeatherConfig:
    """Configuration for weather service."""
    api_key: str
    lat: float
    lon: float
    units: str = "metric"
    location: Optional[str] = None  # Made location optional
    # Allow for additional fields without causing errors
    def __init__(self, **kwargs):
        self.api_key = kwargs.get('api_key')
        self.lat = kwargs.get('lat')
        self.lon = kwargs.get('lon')
        self.units = kwargs.get('units', 'metric')
        self.location = kwargs.get('location')
        # Store any additional config fields
        self.additional_config = {k: v for k, v in kwargs.items()
                                if k not in ['api_key', 'lat', 'lon', 'units', 'location']}

class WeatherService:
    """Service for fetching weather data from OpenWeatherMap API."""

    BASE_URL = "https://api.openweathermap.org/data/3.0"

    def __init__(self, config_file: str):
        """Initialize the weather service with the given configuration file."""
        self.config = self._load_config(config_file)
        self.client = httpx.AsyncClient(timeout=30.0)

    def _load_config(self, config_file: str) -> WeatherConfig:
        """Loads and validates weather configuration."""
        try:
            with open(config_file, 'r', encoding='utf-8') as f:
                config_data = json.load(f)
                weather_config = config_data.get('weather', {})

                required_fields = ['api_key', 'lat', 'lon']
                missing_fields = [field for field in required_fields
                                if field not in weather_config]

                if missing_fields:
                    raise ValueError(f"Missing required config fields: {missing_fields}")

                return WeatherConfig(**weather_config)
        except FileNotFoundError:
            raise HTTPException(
                status_code=500,
                detail=f"Weather configuration file not found: {config_file}"
            )
        except json.JSONDecodeError:
            raise HTTPException(
                status_code=500,
                detail="Invalid JSON in configuration file"
            )

    async def _make_request(self, endpoint: str, params: Dict) -> Dict:
        """Make HTTP request to OpenWeatherMap API."""
        try:
            response = await self.client.get(
                f"{self.BASE_URL}/{endpoint}",
                params=params
            )

            if response.status_code == 401:
                raise HTTPException(
                    status_code=401,
                    detail="Invalid API key"
                )
            if response.status_code == 429:
                raise HTTPException(
                    status_code=429,
                    detail="API rate limit exceeded"
                )
            if response.status_code != 200:
                raise HTTPException(
                    status_code=response.status_code,
                    detail=f"OpenWeatherMap API error: {response.text}"
                )

            return response.json()
        except httpx.TimeoutException:
            raise HTTPException(
                status_code=504,
                detail="Request to weather service timed out"
            )
        except httpx.RequestError as e:
            raise HTTPException(
                status_code=502,
                detail=f"Failed to connect to weather service: {str(e)}"
            )

    def _parse_hourly_forecast(self, hourly_data: List[Dict], hours: int = 24) -> List[Dict]:
        """Parse hourly forecast data."""
        return [{
            "timestamp": hour['dt'],
            "temperature": {
                "temp": hour['temp'],
                "feels_like": hour['feels_like']
            },
            "status": hour['weather'][0]['main'],
            "icon": hour['weather'][0]['id'],
            "humidity": hour['humidity'],
            "wind_speed": hour['wind_speed']
        } for hour in hourly_data[:hours]]

    def _parse_daily_forecast(self, daily_data: List[Dict], days: int = 7) -> List[Dict]:
        """Parse daily forecast data."""
        return [{
            "timestamp": day['dt'],
            "temperature": {
                "temp": day['temp']['day'],
                "min": day['temp']['min'],
                "max": day['temp']['max'],
                "feels_like": day['feels_like']['day']
            },
            "status": day['weather'][0]['main'],
            "icon": day['weather'][0]['id']
        } for day in daily_data[:days]]

    async def get_weather(self) -> Dict:
        """Fetches the current weather and forecasts."""
        params = {
            "lat": self.config.lat,
            "lon": self.config.lon,
            "appid": self.config.api_key,
            "units": self.config.units,
            "exclude": "minutely,alerts"
        }

        data = await self._make_request("onecall", params)

        current = data['current']
        daily = data['daily'][0]

        response = {
            "latitude": data['lat'],
            "longitude": data['lon'],
            "temperature": {
                "temp": current['temp'],
                "min": daily['temp']['min'],
                "max": daily['temp']['max'],
                "feels_like": current['feels_like']
            },
            "status": current['weather'][0]['main'],
            "detailed_status": current['weather'][0]['description'],
            "icon": current['weather'][0]['id'],
            "humidity": current['humidity'],
            "wind_speed": current['wind_speed'],
            "wind_dir": current['wind_deg'],
            "uvi": current['uvi'],
            "clouds": current['clouds'],
            "summary": daily.get('summary', ''),
            "hourly_forecast": self._parse_hourly_forecast(data['hourly']),
            "daily_forecast": self._parse_daily_forecast(data['daily'])
        }

        # Add location if available
        if self.config.location:
            response["location"] = self.config.location

        return response

    async def __aenter__(self):
        """Async context manager entry."""
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self.client.aclose()