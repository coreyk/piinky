package calendarservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/coreyk/piinky/backend-go/models"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Service struct {
	calendarSvc    *calendar.Service
	calendarConfig models.CalendarConfig
	now            func() time.Time // For testing purposes
}

func NewService() (*Service, error) {
	// Get the calendar config filename from the environment variable
	calendarConfigFile := os.Getenv("GOOGLE_CALENDAR_CONFIG_FILE")
	if calendarConfigFile == "" {
		calendarConfigFile = "../google_calendar_config.json"
		fmt.Println("GOOGLE_CALENDAR_CONFIG_FILE is not set in the .env file, using default: ", calendarConfigFile)
	}

	calendarConfigBytes, err := os.ReadFile(calendarConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read calendar config: %v", err)
	}

	var calendarConfig models.CalendarConfig
	if err := json.Unmarshal(calendarConfigBytes, &calendarConfig); err != nil {
		return nil, fmt.Errorf("failed to parse calendar config: %v", err)
	}

	calendarCredentialsFile := os.Getenv("GOOGLE_CALENDAR_CREDENTIALS_FILE")
	if calendarCredentialsFile == "" {
		calendarCredentialsFile = "../google_credentials.json"
		fmt.Println("GOOGLE_CALENDAR_CREDENTIALS_FILE is not set in the .env file, using default: ", calendarCredentialsFile)
	}

	calendarCredentialsBytes, err := os.ReadFile(calendarCredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials: %v", err)
	}

	calendarCreds, err := google.JWTConfigFromJSON(calendarCredentialsBytes, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %v", err)
	}

	calendarSvc, err := calendar.NewService(context.Background(), option.WithHTTPClient(calendarCreds.Client(context.Background())))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %v", err)
	}

	return &Service{
		calendarSvc:    calendarSvc,
		calendarConfig: calendarConfig,
		now:            time.Now,
	}, nil
}

// SetNowFunc allows setting a custom time function for testing
func (s *Service) SetNowFunc(f func() time.Time) {
	s.now = f
}

func (s *Service) HandleGetCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current time in local timezone
	today := s.now().In(time.Local)

	// Calculate the start of the week based on the configuration
	var weekStart time.Time
	if s.calendarConfig.StartOnSunday {
		// If today is Sunday (0), we want to stay on this Sunday
		daysToSubtract := int(today.Weekday())
		if daysToSubtract == 0 {
			weekStart = today
		} else {
			weekStart = today.AddDate(0, 0, -daysToSubtract)
		}
	} else {
		// For Monday start: if today is Monday (1), we stay on Monday
		daysToSubtract := int(today.Weekday())
		if daysToSubtract == 0 { // Sunday
			daysToSubtract = 6
		} else {
			daysToSubtract--
		}
		weekStart = today.AddDate(0, 0, -daysToSubtract)
	}

	// Ensure we start at beginning of day in local time
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.Local)

	timeMin := weekStart.Format(time.RFC3339)
	timeMax := today.AddDate(0, 0, s.calendarConfig.NumberOfWeeks*7).Format(time.RFC3339)

	var events []*calendar.Event
	for _, cal := range s.calendarConfig.Calendars {
		result, err := s.calendarSvc.Events.List(cal.CalendarID).
			TimeMin(timeMin).
			TimeMax(timeMax).
			SingleEvents(true).
			OrderBy("startTime").
			Do()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch calendar events: %v", err), http.StatusInternalServerError)
			return
		}

		for _, event := range result.Items {
			event.ColorId = cal.Color
			events = append(events, event)
		}
	}

	response := map[string]interface{}{
		"startDate":     timeMin,
		"endDate":       timeMax,
		"numberOfWeeks": s.calendarConfig.NumberOfWeeks,
		"startOnSunday": s.calendarConfig.StartOnSunday,
		"events":        events,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
