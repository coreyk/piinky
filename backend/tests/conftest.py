"""
Test configuration and shared fixtures.
"""

import os
import pytest

@pytest.fixture(autouse=True)
def setup_test_env():
    """Set up test environment variables."""
    os.environ['GOOGLE_CALENDAR_CREDENTIALS_FILE'] = 'test_calendar_creds.json'
    os.environ['GOOGLE_CALENDAR_CONFIG_FILE'] = 'test_calendar_config.json'
    os.environ['OWM_CONFIG_FILE'] = 'test_weather_config.json'
    yield
    # Clean up
    os.environ.pop('GOOGLE_CALENDAR_CREDENTIALS_FILE', None)
    os.environ.pop('GOOGLE_CALENDAR_CONFIG_FILE', None)
    os.environ.pop('OWM_CONFIG_FILE', None)