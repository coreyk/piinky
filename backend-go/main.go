package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/coreyk/piinky/backend-go/calendarservice"
	"github.com/coreyk/piinky/backend-go/weatherservice"
	"github.com/joho/godotenv"
)

type Server struct {
	mux         *http.ServeMux
	calendarSvc *calendarservice.Service
	weatherSvc  *weatherservice.Service
}

func NewServer() (*Server, error) {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %v", err)
	}

	calendarSvc, err := calendarservice.NewService()
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %v", err)
	}

	weatherSvc, err := weatherservice.NewService()
	if err != nil {
		return nil, fmt.Errorf("failed to create weather service: %v", err)
	}

	return &Server{
		mux:         http.NewServeMux(),
		calendarSvc: calendarSvc,
		weatherSvc:  weatherSvc,
	}, nil
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("GET /api/calendar", s.cors(s.calendarSvc.HandleGetCalendar))
	s.mux.HandleFunc("GET /api/weather", s.cors(s.weatherSvc.HandleGetWeather))
}

func (s *Server) cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	server.setupRoutes()

	log.Printf("Server starting on :8000")
	log.Fatal(http.ListenAndServe(":8000", server.mux))
}
