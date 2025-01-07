"""
This module initializes the FastAPI application, sets up middleware, and defines
API endpoints for calendar and weather services.
"""

import os
from fastapi import FastAPI, Response
from fastapi.middleware.cors import CORSMiddleware
from dotenv import load_dotenv
from piinky_backend.calendar_service import CalendarService
from piinky_backend.weather_service import WeatherService

# Load environment variables
load_dotenv()

app = FastAPI()

# Configure CORS
origins = ["*"]
app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["GET", "OPTIONS"],
    allow_headers=["*"],
    expose_headers=["*"],
)

DEVELOPMENT_MODE = os.getenv('DEVELOPMENT_MODE', 'false').lower() == 'true'
GOOGLE_CALENDAR_CREDENTIALS_FILE = os.getenv(
    'GOOGLE_CALENDAR_CREDENTIALS_FILE', '../google_credentials.json'
)
GOOGLE_CALENDAR_CONFIG_FILE = os.getenv(
    'GOOGLE_CALENDAR_CONFIG_FILE', '../google_calendar_config.json'
)
OWM_CONFIG_FILE = os.getenv('OWM_CONFIG_FILE', '../owm_config.json')
calendar_service = CalendarService(GOOGLE_CALENDAR_CREDENTIALS_FILE, GOOGLE_CALENDAR_CONFIG_FILE)
weather_service = WeatherService(OWM_CONFIG_FILE)

@app.options("/api/calendar")
@app.options("/api/weather")
async def options_handler():
    """Handle OPTIONS requests."""
    response = Response()
    response.headers["Access-Control-Allow-Origin"] = "*"
    response.headers["Access-Control-Allow-Methods"] = "GET, OPTIONS"
    return response

@app.get("/api/calendar")
async def get_calendar():
    """
    Retrieve calendar data from the calendar service.
    """
    return await calendar_service.get_calendar_data()

@app.get("/api/weather")
async def get_weather():
    """
    Retrieve weather data from the weather service.
    """
    return await weather_service.get_weather()
