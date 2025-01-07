"""
Tests for the weather service.
"""

import json
import os
from unittest.mock import MagicMock, AsyncMock, patch
import pytest
from fastapi import HTTPException
import httpx
from piinky_backend.weather_service import WeatherService, WeatherConfig

@pytest.fixture
def mock_config():
    """
    Fixture providing mock weather configuration.
    """
    return {
        "weather": {
            "api_key": "test-api-key",
            "lat": 37.7749,
            "lon": -122.4194,
            "location": "San Francisco",
            "units": "metric"
        }
    }

@pytest.fixture
def mock_service(mock_config, tmp_path):
    """
    Fixture providing a WeatherService instance with mocked dependencies.
    """
    # Create temporary config file in a temporary directory
    config_file = tmp_path / "test_weather_config.json"
    with open(config_file, "w", encoding="utf-8") as f:
        json.dump(mock_config, f)

    service = WeatherService(str(config_file))

    # Create a mock client with async methods
    mock_client = AsyncMock()
    mock_client.get = AsyncMock()
    service.client = mock_client

    yield service

@pytest.fixture
def mock_weather_response():
    """
    Fixture providing a mock OpenWeatherMap API response.
    """
    return {
        "lat": 37.7749,
        "lon": -122.4194,
        "current": {
            "dt": 1641456000,
            "sunrise": 1641484400,
            "sunset": 1641520800,
            "temp": 20.5,
            "feels_like": 19.8,
            "humidity": 65,
            "dew_point": 13.8,
            "uvi": 4.5,
            "clouds": 75,
            "visibility": 10000,
            "wind_speed": 3.5,
            "wind_deg": 180,
            "weather": [
                {
                    "id": 800,
                    "main": "Clear",
                    "description": "clear sky",
                    "icon": "01d"
                }
            ]
        },
        "hourly": [
            {
                "dt": 1641456000,
                "temp": 20.5,
                "feels_like": 19.8,
                "humidity": 65,
                "wind_speed": 3.5,
                "weather": [
                    {
                        "id": 800,
                        "main": "Clear",
                        "description": "clear sky",
                        "icon": "01d"
                    }
                ]
            }
        ],
        "daily": [
            {
                "dt": 1641456000,
                "temp": {
                    "day": 20.5,
                    "min": 15.0,
                    "max": 25.0,
                    "night": 15.0,
                    "eve": 18.0,
                    "morn": 16.0
                },
                "feels_like": {
                    "day": 19.8,
                    "night": 14.5,
                    "eve": 17.5,
                    "morn": 15.5
                },
                "weather": [
                    {
                        "id": 800,
                        "main": "Clear",
                        "description": "clear sky",
                        "icon": "01d"
                    }
                ],
                "summary": "Clear day"
            }
        ]
    }

def test_weather_config():
    """Test WeatherConfig initialization with additional fields."""
    config = WeatherConfig(
        api_key="test-key",
        lat=37.7749,
        lon=-122.4194,
        units="metric",
        location="Test City",
        extra_field="value"
    )
    assert config.api_key == "test-key"
    assert config.lat == 37.7749
    assert config.lon == -122.4194
    assert config.units == "metric"
    assert config.location == "Test City"
    assert config.additional_config == {"extra_field": "value"}

@pytest.mark.asyncio
async def test_get_weather(mock_service, mock_weather_response):
    """
    Test retrieving weather data.
    """
    # Mock the API response
    mock_response = AsyncMock()
    mock_response.status_code = 200
    mock_response.json = MagicMock(return_value=mock_weather_response)
    mock_service.client.get.return_value = mock_response

    result = await mock_service.get_weather()

    # Verify API call
    mock_service.client.get.assert_called_once_with(
        "https://api.openweathermap.org/data/3.0/onecall",
        params={
            "lat": 37.7749,
            "lon": -122.4194,
            "appid": "test-api-key",
            "units": "metric",
            "exclude": "minutely,alerts"
        }
    )

    assert result['latitude'] == 37.7749
    assert result['longitude'] == -122.4194
    assert result['temperature']['temp'] == 20.5
    assert result['temperature']['min'] == 15.0
    assert result['temperature']['max'] == 25.0
    assert result['temperature']['feels_like'] == 19.8
    assert result['status'] == 'Clear'
    assert result['detailed_status'] == 'clear sky'
    assert result['icon'] == 800
    assert result['humidity'] == 65
    assert result['wind_speed'] == 3.5
    assert result['wind_dir'] == 180
    assert result['uvi'] == 4.5
    assert result['clouds'] == 75
    assert result['summary'] == 'Clear day'
    assert result['location'] == 'San Francisco'
    assert len(result['hourly_forecast']) == 1
    assert len(result['daily_forecast']) == 1

@pytest.mark.asyncio
async def test_get_weather_unauthorized(mock_service):
    """
    Test error handling when the API key is invalid.
    """
    mock_response = AsyncMock()
    mock_response.status_code = 401
    mock_response.text = "Invalid API key"
    mock_service.client.get.return_value = mock_response

    with pytest.raises(HTTPException) as exc_info:
        await mock_service.get_weather()

    assert exc_info.value.status_code == 401
    assert exc_info.value.detail == "Invalid API key"

@pytest.mark.asyncio
async def test_get_weather_rate_limit(mock_service):
    """
    Test error handling when API rate limit is exceeded.
    """
    mock_response = AsyncMock()
    mock_response.status_code = 429
    mock_response.text = "Rate limit exceeded"
    mock_service.client.get.return_value = mock_response

    with pytest.raises(HTTPException) as exc_info:
        await mock_service.get_weather()

    assert exc_info.value.status_code == 429
    assert exc_info.value.detail == "API rate limit exceeded"

@pytest.mark.asyncio
async def test_get_weather_timeout(mock_service):
    """
    Test error handling when the API request times out.
    """
    mock_service.client.get.side_effect = httpx.TimeoutException("Request timed out")

    with pytest.raises(HTTPException) as exc_info:
        await mock_service.get_weather()

    assert exc_info.value.status_code == 504
    assert exc_info.value.detail == "Request to weather service timed out"

@pytest.mark.asyncio
async def test_get_weather_connection_error(mock_service):
    """
    Test error handling when there's a connection error.
    """
    mock_service.client.get.side_effect = httpx.RequestError("Connection failed")

    with pytest.raises(HTTPException) as exc_info:
        await mock_service.get_weather()

    assert exc_info.value.status_code == 502
    assert "Failed to connect to weather service" in exc_info.value.detail

@pytest.mark.asyncio
async def test_missing_config_file():
    """
    Test error handling when config file is missing.
    """
    with pytest.raises(HTTPException) as exc_info:
        WeatherService("nonexistent_config.json")

    assert exc_info.value.status_code == 500
    assert "Weather configuration file not found" in exc_info.value.detail

@pytest.mark.asyncio
async def test_invalid_config_json(tmp_path):
    """
    Test error handling when config file contains invalid JSON.
    """
    config_file = tmp_path / "invalid_config.json"
    with open(config_file, "w") as f:
        f.write("invalid json")

    with pytest.raises(HTTPException) as exc_info:
        WeatherService(str(config_file))

    assert exc_info.value.status_code == 500
    assert "Invalid JSON in configuration file" == exc_info.value.detail