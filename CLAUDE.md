# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Piinky is a Raspberry Pi-powered e-ink calendar display for the fridge. It shows a multi-week calendar with Google Calendar events and current weather information on a Pimoroni Inky Impression 7.3" display (800x480 resolution).

## Architecture

The project has three main components that run together:

1. **Go Backend** (`backend-go/`) - API server on port 8000
   - `calendarservice/` - Fetches events from Google Calendar API using service account credentials
   - `weatherservice/` - Fetches weather from OpenWeatherMap 3.0 API
   - `models/` - Shared data structures
   - Endpoints: `GET /api/calendar`, `GET /api/weather`

2. **React Frontend** (`frontend/`) - Vite dev server on port 3000
   - `components/Calendar.tsx` - Main calendar grid component
   - `components/CalendarDay.tsx` - Individual day cell with events
   - `components/Weather.tsx` - Weather display widget
   - `components/Event.tsx` - Event rendering
   - `lib/colormap.ts` - Calendar color definitions

3. **Python Display Service** (`display/`) - Playwright-based screenshot service
   - Takes screenshots of the frontend every 4 hours
   - Renders to the Inky Impression e-ink display via SPI
   - Uses `inky` library (only on Raspberry Pi hardware)

## Development Commands

### Run all services together
```bash
./piinky.sh              # Default: Go backend + Python display
./piinky.sh go python    # Explicit: Go backend + Python display
```

### Frontend (from `frontend/`)
```bash
npm install
npm run dev              # Start dev server on port 3000
npm test                 # Run tests with vitest
npm run test:watch       # Watch mode
npm run test:coverage    # Coverage report
npm run build            # Production build
```

### Go Backend (from `backend-go/`)
```bash
./piinky.sh --create-env backend-go    # Create .env file first
go mod tidy
go run main.go                         # Run server
go test ./... -v                       # Run all tests
air                                    # Hot reload (install: go install github.com/air-verse/air@latest)
```

### Python Display Service (from `display/`)
```bash
python3 -m venv venv && source venv/bin/activate
pip install -e .                       # Basic install
pip install -e ".[pi]"                 # With Raspberry Pi dependencies
pip install -e ".[test]"               # With test dependencies
pytest -v
```

### Python Backend (alternative, from `backend/`)
```bash
./piinky.sh --create-env backend
pip install -e .
pip install -e ".[test]"
pytest -v
pytest --cov=. tests/                  # Coverage
```

## Configuration Files

All config files live in the project root:
- `google_credentials.json` - Google service account key (see `.example`)
- `google_calendar_config.json` - Calendar IDs and colors (see `.example`)
- `owm_config.json` - OpenWeatherMap API key and location (see `.example`)

Environment variables (`.env` files in each backend) reference these:
- `GOOGLE_CALENDAR_CREDENTIALS_FILE`
- `GOOGLE_CALENDAR_CONFIG_FILE`
- `OWM_CONFIG_FILE`
- `DEVELOPMENT_MODE`

## Hardware Context

- Target: Raspberry Pi 5 with Pimoroni Inky Impression 7.3"
- Display resolution: 800x480 pixels
- The display service auto-detects Pi (aarch64 Linux) vs development mode
- SPI must be enabled on the Pi (`dtoverlay=spi0-0cs` in `/boot/firmware/config.txt`)
