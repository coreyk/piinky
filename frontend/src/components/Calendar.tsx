import React, { useState, useEffect } from 'react';
import './Calendar.scss';
import CalendarDay from './CalendarDay';
import { CalendarData, DayData, CalendarEvent } from '../types/calendar';
import { ForecastData } from '../types/weather';
import Weather from './Weather';

const Calendar: React.FC<{ headerPosition?: 'top' | 'bottom' }> = ({ headerPosition = 'bottom' }) => {
  const API_HOST = import.meta.env.VITE_API_HOST || window.location.protocol + '//' + window.location.hostname;
  const API_PORT = import.meta.env.VITE_API_PORT || '8000';
  const API_URL = `${API_HOST}:${API_PORT}`;
  const [calendarData, setCalendarData] = useState<CalendarData | null>(null);
  const [dailyForecast, setDailyForecast] = useState<ForecastData[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchWeather = async () => {
      try {
        const response = await fetch(`${API_URL}/api/weather`);
        if (!response.ok) {
          throw new Error('Failed to fetch weather data');
        }
        const weatherData = await response.json();
        setDailyForecast(weatherData.daily_forecast);
      } catch (error) {
        console.error('Error fetching weather:', error);
      }
    };

    let weatherMounted = true;
    if (weatherMounted) {
      fetchWeather();
    }

    const weatherInterval = setInterval(fetchWeather, 300000);

    return () => {
      weatherMounted = false;
      clearInterval(weatherInterval);
    };
  }, []);

  useEffect(() => {
    const fetchCalendar = async () => {
      try {
        const response = await fetch(`${API_URL}/api/calendar`);
        if (!response.ok) {
          throw new Error('Failed to fetch calendar data');
        }
        const { startDate, startOnSunday, numberOfWeeks, events } = await response.json();

        const start = new Date(startDate);
        const currentWeekStart = new Date(start);
        // Reset time to midnight so all-day events are displayed correctly
        currentWeekStart.setHours(0, 0, 0, 0);

        const days: DayData[] = Array.from({ length: numberOfWeeks * 7 }, (_, index) => {
          const date = new Date(currentWeekStart);
          date.setDate(currentWeekStart.getDate() + index);
          return {
            date: date.toISOString(),
            isCurrentMonth: date.getMonth() === new Date().getMonth(),
            isToday: date.toDateString() === new Date().toDateString(),
            events: events.filter((event: CalendarEvent) => {
              const startDateTime = event.start.dateTime || event.start.date;
              const endDateTime = event.end.dateTime || event.end.date;
              if (!startDateTime || !endDateTime) return false;

              if (event.start.date) {
                // For date-only (all-day) events and multi-day events
                const eventStart = new Date(event.start.date + 'T00:00:00');
                const eventEnd = new Date(event.end.date + 'T00:00:00');
                const compareDate = new Date(date.toISOString().split('T')[0] + 'T00:00:00');

                // Subtract one day from end date because Google Calendar's end date is exclusive
                eventEnd.setDate(eventEnd.getDate() - 1);

                return compareDate >= eventStart && compareDate <= eventEnd;
              } else {
                const eventDate = new Date(startDateTime);
                return eventDate.toDateString() === date.toDateString();
              }
            })
          };
        });

        setCalendarData({
          startDate: startDate,
          endDate: new Date(start.getFullYear(), start.getMonth() + 6, 0).toISOString(),
          startOnSunday: startOnSunday,
          numberOfWeeks: numberOfWeeks,
          events: days
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      }
    };

    let mounted = true;
    if (mounted) {
      fetchCalendar();
    }

    const interval = setInterval(fetchCalendar, 5 * 60 * 1000);

    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, []);

  if (error) {
    return <div>Error: {error}</div>;
  }

  if (!calendarData) {
    return <div>Loading...</div>;
  }

  const currentDate = new Date(Date.now());
  const monthYear = currentDate.toLocaleDateString('en-US', {
    month: 'short',
    year: 'numeric'
  }).toUpperCase();
  const daysOfWeek = calendarData.startOnSunday ? ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'] : ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

  const renderHeader = () => (
    <div className="col-span-7 bg-white px-2 py-1 flex justify-between items-center">
      <h2 className="text-md font-bold text-black-800 mb-2">{monthYear}</h2>
      <Weather />
    </div>
  );

  return (
    <div className="max-w-6xl mx-auto bg-white overflow-hidden">
      {headerPosition === 'top' && renderHeader()}
      <div className="grid grid-cols-7 gap-px bg-gray-500 border border-blue-500">
        {daysOfWeek.map((day, index) => (
          <div
            key={`header-${index}`}
            className="bg-white p-0 text-center text-sm font-semibold text-black-800 border-r border-b border-t border-blue-500"
          >
            {day}
          </div>
        ))}
        {calendarData.events.map((day, index) => (
          <CalendarDay
            key={`day-${index}`}
            dayData={day}
            calendarData={calendarData}
            dailyForecast={dailyForecast}
          />
        ))}
      </div>
      {headerPosition === 'bottom' && renderHeader()}
    </div>
  );
};

export default Calendar;