#!/bin/bash

# Defaults
BACKEND="go"
DISPLAY="python"

# Function to create .env files
create_env_files() {
    local backend_dir=$1

    if [ ! -f "$backend_dir/.env" ]; then
        echo "Creating $backend_dir .env file..."
        echo "GOOGLE_CALENDAR_CREDENTIALS_FILE=../google_credentials.json" > "$backend_dir/.env"
        echo "GOOGLE_CALENDAR_CONFIG_FILE=../google_calendar_config.json" >> "$backend_dir/.env"
        echo "OWM_CONFIG_FILE=../owm_config.json" >> "$backend_dir/.env"
        echo "DEVELOPMENT_MODE=true" >> "$backend_dir/.env"
    fi
}

# Check for command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --create-env)
            if [ -z "$2" ]; then
                echo "Error: --create-env requires a backend directory (backend or backend-go) as an argument."
                exit 1
            fi
            create_env_files "$2"
            exit 0
            ;;
        python|go)
            BACKEND="$1"
            shift
            if [[ $# -gt 0 && ("$1" == "python" || "$1" == "go") ]]; then
                DISPLAY="$1"
                shift
            else
                DISPLAY="python"
            fi
            ;;
        *)
            echo "Unknown argument: $1"
            exit 1
            ;;
    esac
done

echo "Starting piinky $BACKEND backend with $DISPLAY display..."

# Create .env files for both backend directories if they don't exist
create_env_files "$(dirname "$0")/backend"
create_env_files "$(dirname "$0")/backend-go"

# Backend setup
if [ "$BACKEND" == "python" ]; then
    cd backend
    python3 -m venv venv
    source venv/bin/activate

    echo "Installing Python backend dependencies..."
    pip install -e .
    # Check if we're on Raspberry Pi
    if [[ $(uname -s) == Linux && $(uname -m) == aarch* ]]; then
        echo "Installing Raspberry Pi-specific backend dependencies..."
        pip install -e ".[pi]"
    fi

    # uvicorn main:app --reload --host 0.0.0.0 --port 8000 &
    python -m uvicorn piinky_backend.main:app --reload --port 8000 &
    BACKEND_PID=$!

elif [ "$BACKEND" == "go" ]; then
    cd backend-go
    echo "Installing Go backend dependencies..."

    go mod tidy

    if [[ $(uname -s) == Linux && $(uname -m) == aarch* ]]; then
        go run main.go &
    else
        go install github.com/air-verse/air@latest
        air -d &
    fi
    BACKEND_PID=$!
fi

echo "Starting piinky frontend..."
cd ../frontend
echo "Installing frontend dependencies..."
npm install
if [ ! -f .env ]; then
    echo "Creating frontend .env file..."
    echo "VITE_API_HOST=http://localhost" > .env
    echo "VITE_API_PORT=8000" >> .env
fi
npm start &
FRONTEND_PID=$!


if [ "$DISPLAY" == "python" ]; then
    echo "Starting Python display service..."
    cd ../display
    python3 -m venv venv
    source venv/bin/activate
    echo "Installing Python display dependencies..."
    pip install -e .
    # Check if we're on Raspberry Pi
    if [[ $(uname -s) == Linux && $(uname -m) == aarch* ]]; then
        echo "Installing Raspberry Pi-specific display dependencies..."
        pip install -e ".[pi]"
    fi
    python main.py &
    DISPLAY_PID=$!

elif [ "$DISPLAY" == "go" ]; then
    echo "Starting Go display service..."
    cd ../display-go
    if [[ $(uname -s) == Linux && $(uname -m) == aarch* ]]; then
        go run main.go &
    else
        air -d &
    fi
    DISPLAY_PID=$!
fi

# Handle shutdown
trap "kill $BACKEND_PID $FRONTEND_PID $DISPLAY_PID" SIGINT SIGTERM EXIT

# Keep script running
wait