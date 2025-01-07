"""
Tests for the main FastAPI application.
"""

import json
import os
from unittest.mock import patch
import pytest
from fastapi.testclient import TestClient
from piinky_backend.main import app

@pytest.fixture
def test_client():
    """Create a test client for the FastAPI application."""
    return TestClient(app)

@pytest.fixture
def mock_calendar_data():
    """Create mock calendar data."""
    return {
        "startDate": "2024-01-07T00:00:00Z",
        "endDate": "2024-01-21T00:00:00Z",
        "numberOfWeeks": 2,
        "startOnSunday": True,
        "events": [
            {
                "id": "123",
                "summary": "Test Event",
                "start": {"dateTime": "2024-01-07T10:00:00Z"},
                "end": {"dateTime": "2024-01-07T11:00:00Z"},
                "colorId": "#4285F4"
            }
        ]
    }

@pytest.fixture
def mock_weather_data():
    """Create mock weather data."""
    return {
        "lat": 37.7749,
        "lon": -122.4194,
        "current": {
            "dt": 1641456000,
            "temp": 20.5,
            "feels_like": 19.8,
            "humidity": 65,
            "weather": [
                {
                    "id": 800,
                    "main": "Clear",
                    "description": "clear sky",
                    "icon": "01d"
                }
            ]
        },
        "daily": [
            {
                "dt": 1641456000,
                "temp": {
                    "day": 20.5,
                    "min": 15.0,
                    "max": 25.0
                },
                "weather": [
                    {
                        "id": 800,
                        "main": "Clear"
                    }
                ]
            }
        ],
        "hourly": [
            {
                "dt": 1641456000,
                "temp": 20.5,
                "feels_like": 19.8,
                "humidity": 65,
                "weather": [
                    {
                        "id": 800,
                        "main": "Clear"
                    }
                ]
            }
        ]
    }

def test_cors_headers(test_client):
    """Test that CORS headers are properly set."""
    response = test_client.options("/api/calendar")
    assert response.status_code == 200
    assert response.headers["access-control-allow-origin"] == "*"
    assert "GET" in response.headers["access-control-allow-methods"]

@pytest.mark.asyncio
async def test_get_calendar(test_client, mock_calendar_data):
    """Test the calendar endpoint."""
    with patch('piinky_backend.main.calendar_service.get_calendar_data') as mock_get_calendar:
        mock_get_calendar.return_value = mock_calendar_data
        response = test_client.get("/api/calendar")

        assert response.status_code == 200
        data = response.json()
        assert data["numberOfWeeks"] == 2
        assert data["startOnSunday"] is True
        assert len(data["events"]) == 1

@pytest.mark.asyncio
async def test_get_weather(test_client, mock_weather_data):
    """Test the weather endpoint."""
    with patch('piinky_backend.main.weather_service.get_weather') as mock_get_weather:
        mock_get_weather.return_value = mock_weather_data
        response = test_client.get("/api/weather")

        assert response.status_code == 200
        data = response.json()
        assert data["lat"] == 37.7749
        assert data["lon"] == -122.4194
        assert data["current"]["temp"] == 20.5

def test_environment_variables():
    """Test that environment variables are properly loaded."""
    assert os.getenv('GOOGLE_CALENDAR_CREDENTIALS_FILE') is not None
    assert os.getenv('GOOGLE_CALENDAR_CONFIG_FILE') is not None
    assert os.getenv('OWM_CONFIG_FILE') is not None