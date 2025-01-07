import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import Calendar from './Calendar';

// Mock the Weather component
vi.mock('./Weather', () => ({
  default: () => <div data-testid="weather-mock">Weather Widget</div>
}));

// Mock calendar and weather data
const mockCalendarData = {
  startDate: '2024-01-15T00:00:00.000Z', // Match our test date
  startOnSunday: true,
  numberOfWeeks: 2,
  events: []
};

const mockWeatherData = {
  daily_forecast: [],
  hourly_forecast: [],
  temperature: {
    temp: 72,
    feels_like: 70,
    min: 65,
    max: 75
  },
  status: 'Clear',
  icon: '800',
  humidity: 45,
  wind_speed: 8,
  summary: 'Clear skies'
};

describe('Calendar', () => {
  const originalConsoleError = console.error;

  beforeEach(() => {
    vi.clearAllMocks();
    // Mock Date.now to return January 15, 2024
    vi.spyOn(Date, 'now').mockImplementation(() => new Date('2024-01-15T00:00:00.000Z').getTime());
    // Mock console.error to suppress expected error messages
    console.error = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    console.error = originalConsoleError;
  });

  it('renders loading state initially', async () => {
    // Create a promise that we won't resolve immediately
    let resolveCalendarPromise!: (value: any) => void;
    let resolveWeatherPromise!: (value: any) => void;
    const calendarPromise = new Promise(resolve => { resolveCalendarPromise = resolve; });
    const weatherPromise = new Promise(resolve => { resolveWeatherPromise = resolve; });

    // Setup fetch mock to return our unresolved promises
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes('/api/calendar')) {
        return Promise.resolve({
          ok: true,
          json: () => calendarPromise
        });
      } else if (url.includes('/api/weather')) {
        return Promise.resolve({
          ok: true,
          json: () => weatherPromise
        });
      }
      return Promise.reject(new Error('Not found'));
    });

    render(<Calendar />);

    // Check for loading state before resolving promises
    expect(screen.getByText('Loading...')).toBeInTheDocument();

    // Now resolve the promises
    resolveCalendarPromise(mockCalendarData);
    resolveWeatherPromise(mockWeatherData);

    // Wait for loading to disappear
    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });
  });

  it('renders calendar after data is loaded', async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes('/api/calendar')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockCalendarData),
        });
      } else if (url.includes('/api/weather')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockWeatherData),
        });
      }
      return Promise.reject(new Error('Not found'));
    });

    await act(async () => {
      render(<Calendar />);
    });

    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    // Check if days of week are rendered
    const daysOfWeek = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
    daysOfWeek.forEach(day => {
      expect(screen.getByText(day)).toBeInTheDocument();
    });
  });

  it('renders with correct header position', async () => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url.includes('/api/calendar')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockCalendarData),
        });
      } else if (url.includes('/api/weather')) {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockWeatherData),
        });
      }
      return Promise.reject(new Error('Not found'));
    });

    await act(async () => {
      render(<Calendar headerPosition="top" />);
    });

    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    // Check if the header is rendered with the correct class
    const header = screen.getByRole('heading', { level: 2 }).closest('.col-span-7');
    expect(header).toBeInTheDocument();

    // Check if the current month and year are displayed
    const monthYear = 'JAN 2024';
    expect(screen.getByText(monthYear)).toBeInTheDocument();
  });

  it('handles fetch error gracefully', async () => {
    // Mock a failed fetch for both endpoints
    global.fetch = vi.fn().mockRejectedValue(new Error('Failed to fetch'));

    await act(async () => {
      render(<Calendar />);
    });

    await waitFor(() => {
      expect(screen.getByText('Error: Failed to fetch')).toBeInTheDocument();
      expect(console.error).toHaveBeenCalled();
    });
  });
});