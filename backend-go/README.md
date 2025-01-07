# Piinky Backend (Go)

This is a Go implementation of the Piinky backend service, designed to be a drop-in replacement for the Python version.

## Features

- E-ink display support using a modified version of the Inky driver
- Google Calendar integration
- OpenWeatherMap integration
- Web screenshot functionality using gowitness
- RESTful API using Fiber

## Prerequisites

- Go 1.21 or later
- Chrome/Chromium (for screenshots)
- Raspberry Pi (for e-ink display support)

## Setup

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up configuration files:
   ```bash
   cp config/google_calendar_config.json.example config/google_calendar_config.json
   cp config/owm_config.json.example config/owm_config.json
   ```

4. Edit the configuration files with your API keys and settings:
   - `config/google_calendar_config.json`: Add your Google Calendar IDs and color preferences
   - `config/owm_config.json`: Add your OpenWeatherMap API key
   - Place your Google service account credentials in `config/google_credentials.json`

## Running

```bash
go run main.go
```

The server will start on port 8000 and begin taking screenshots of the frontend (expected to be running on port 3000) every 2 hours.

## API Endpoints

- `/api/calendar`: Get calendar events for the next 3 weeks
- `/api/weather`: Get current weather data

## Notes

- The e-ink display functionality is only active when running on a Raspberry Pi
- When not running on a Pi, the display updates will be simulated
- The screenshot functionality requires Chrome/Chromium to be installed