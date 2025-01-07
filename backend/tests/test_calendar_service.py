"""
Tests for the calendar service.
"""

import json
import os
from datetime import datetime, timedelta
from unittest.mock import MagicMock, patch
import pytest
from fastapi import HTTPException
from piinky_backend.calendar_service import CalendarService

@pytest.fixture
def mock_config():
    """
    Fixture providing mock calendar configuration.
    """
    return {
        "calendars": [
            {
                "calendar_id": "primary",
                "color": "#4285F4"
            }
        ],
        "start_on_sunday": True,
        "number_of_weeks": 2
    }

@pytest.fixture
def mock_credentials():
    """
    Fixture providing mock Google Calendar credentials.
    """
    return {
        "type": "service_account",
        "project_id": "test-project",
        "private_key_id": "test-key-id",
        "private_key": "test-key",
        "client_email": "test@example.com",
        "client_id": "test-client-id",
        "auth_uri": "https://accounts.google.com/o/oauth2/auth",
        "token_uri": "https://oauth2.googleapis.com/token",
        "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
        "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/test%40example.com"
    }

@pytest.fixture
def mock_service(mock_config, mock_credentials):
    """
    Fixture providing a CalendarService instance with mocked dependencies.
    """
    with patch('piinky_backend.calendar_service.Credentials') as mock_creds, \
         patch('piinky_backend.calendar_service.build') as mock_build:

        # Create temporary config files
        config_file = "test_calendar_config.json"
        creds_file = "test_calendar_creds.json"

        with open(config_file, "w", encoding="utf-8") as f:
            json.dump(mock_config, f)
        with open(creds_file, "w", encoding="utf-8") as f:
            json.dump(mock_credentials, f)

        # Create mock calendar service
        mock_calendar = MagicMock()
        mock_build.return_value = mock_calendar

        service = CalendarService(creds_file, config_file)

        # Clean up temporary files
        os.remove(config_file)
        os.remove(creds_file)

        yield service

@pytest.mark.asyncio
async def test_get_calendar_data_sunday_start(mock_service):
    """
    Test retrieving calendar data with Sunday as the start of the week.
    """
    # Mock the current time to a known Wednesday
    test_date = datetime(2024, 1, 10, 12, 0)  # A Wednesday
    with patch('piinky_backend.calendar_service.datetime') as mock_datetime:
        mock_datetime.now.return_value = test_date
        mock_datetime.side_effect = lambda *args, **kw: datetime(*args, **kw)

        # Mock the events response
        mock_events = {
            'items': [
                {
                    'id': '123',
                    'summary': 'Test Event',
                    'start': {'dateTime': '2024-01-10T10:00:00Z'},
                    'end': {'dateTime': '2024-01-10T11:00:00Z'}
                }
            ]
        }
        mock_service.service.events().list().execute.return_value = mock_events

        result = await mock_service.get_calendar_data()

        assert result['startOnSunday'] is True
        assert result['numberOfWeeks'] == 2

        # Verify the start date is the previous Sunday
        start_date = datetime.fromisoformat(result['startDate'].replace('Z', '+00:00'))
        assert start_date.weekday() == 6  # Sunday
        assert start_date.date() == datetime(2024, 1, 7).date()

@pytest.mark.asyncio
async def test_get_calendar_data_monday_start(mock_service):
    """
    Test retrieving calendar data with Monday as the start of the week.
    """
    mock_service.start_on_sunday = False

    # Mock the current time to a known Thursday
    test_date = datetime(2024, 1, 11, 12, 0)  # A Thursday
    with patch('piinky_backend.calendar_service.datetime') as mock_datetime:
        mock_datetime.now.return_value = test_date
        mock_datetime.side_effect = lambda *args, **kw: datetime(*args, **kw)

        # Mock the events response
        mock_events = {
            'items': [
                {
                    'id': '123',
                    'summary': 'Test Event',
                    'start': {'dateTime': '2024-01-11T10:00:00Z'},
                    'end': {'dateTime': '2024-01-11T11:00:00Z'}
                }
            ]
        }
        mock_service.service.events().list().execute.return_value = mock_events

        result = await mock_service.get_calendar_data()

        assert result['startOnSunday'] is False
        assert result['numberOfWeeks'] == 2

        # Verify the start date is the previous Monday
        start_date = datetime.fromisoformat(result['startDate'].replace('Z', '+00:00'))
        assert start_date.weekday() == 0  # Monday
        assert start_date.date() == datetime(2024, 1, 8).date()

@pytest.mark.asyncio
async def test_get_events(mock_service):
    """
    Test retrieving events from the calendar.
    """
    start_date = datetime.now().isoformat()
    end_date = (datetime.now() + timedelta(days=14)).isoformat()

    # Mock the events response
    mock_events = {
        'items': [
            {
                'id': '123',
                'summary': 'Test Event',
                'start': {'dateTime': '2024-01-10T10:00:00Z'},
                'end': {'dateTime': '2024-01-10T11:00:00Z'}
            }
        ]
    }
    mock_service.service.events().list().execute.return_value = mock_events

    events = await mock_service.get_events(start_date, end_date)

    assert len(events) == 1
    assert events[0]['id'] == '123'
    assert events[0]['colorId'] == '#4285F4'  # From mock config

@pytest.mark.asyncio
async def test_get_events_error_handling(mock_service):
    """
    Test error handling when retrieving events.
    """
    start_date = datetime.now().isoformat()
    end_date = (datetime.now() + timedelta(days=14)).isoformat()

    # Mock an error response
    mock_service.service.events().list().execute.side_effect = Exception("API Error")

    with pytest.raises(HTTPException) as exc_info:
        await mock_service.get_events(start_date, end_date)

    assert exc_info.value.status_code == 500
    assert "API Error" in str(exc_info.value.detail)