import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act } from '@testing-library/react';
import Weather from './Weather';

// Mock weather data
const mockWeatherData = {
  temperature: {
    temp: 72.5,
    max: 75,
    min: 65,
  },
  humidity: "45%",
  wind_speed: 8.5,
  icon: "800",
  summary: "Clear skies",
  hourly_forecast: [
    {
      timestamp: (Date.now() / 1000).toString(),
      temperature: { temp: 72 },
      icon: "800"
    },
    {
      timestamp: (Date.now() / 1000 + 3600).toString(),
      temperature: { temp: 73 },
      icon: "801"
    },
    {
      timestamp: (Date.now() / 1000 + 7200).toString(),
      temperature: { temp: 74 },
      icon: "802"
    }
  ],
  daily_forecast: []
};

describe('Weather', () => {
  const originalConsoleError = console.error;

  beforeEach(() => {
    // Setup default successful fetch mock
    global.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve(mockWeatherData),
      })
    );
    // Mock console.error to suppress expected error messages
    console.error = vi.fn();
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.useRealTimers();
    console.error = originalConsoleError;
  });

  it('renders weather data correctly', async () => {
    await act(async () => {
      render(<Weather />);
    });

    // Wait for the weather data to load
    await waitFor(() => {
      expect(screen.getByText('73°')).toBeInTheDocument();
    });

    // Check temperature display
    expect(screen.getByText('75°')).toBeInTheDocument(); // max temp
    expect(screen.getByText('65°')).toBeInTheDocument(); // min temp

    // Check humidity
    expect(screen.getByText('45%')).toBeInTheDocument();

    // Check wind speed (split into multiple elements)
    const windSpeedElement = screen.getByText('9');
    expect(windSpeedElement).toBeInTheDocument();
    expect(screen.getByText('mph')).toBeInTheDocument();

    // Check weather summary
    expect(screen.getByText('Clear skies')).toBeInTheDocument();
  });

  it('renders hourly forecast', async () => {
    await act(async () => {
      render(<Weather />);
    });

    // Wait for the weather data to load
    await waitFor(() => {
      expect(screen.getByText('72°')).toBeInTheDocument();
    });

    // Check that hourly forecasts are displayed
    mockWeatherData.hourly_forecast.forEach(forecast => {
      const temp = Math.round(forecast.temperature.temp);
      expect(screen.getByText(`${temp}°`)).toBeInTheDocument();
    });
  });

  it('handles fetch error gracefully', async () => {
    // Override the fetch mock for this test only
    global.fetch = vi.fn().mockRejectedValue(new Error('Failed to fetch'));

    await act(async () => {
      render(<Weather />);
    });

    // Component should render nothing when there's an error
    await waitFor(() => {
      expect(document.body.textContent).toBe('');
      expect(console.error).toHaveBeenCalled();
    });
  });

  it('makes API call with correct URL', async () => {
    await act(async () => {
      render(<Weather />);
    });

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('/api/weather'));
    });
  });

  it('updates data periodically', async () => {
    vi.useFakeTimers();

    await act(async () => {
      render(<Weather />);
    });

    // Initial fetch
    expect(global.fetch).toHaveBeenCalledTimes(1);

    // Fast-forward 5 minutes (300000ms)
    await act(async () => {
      await vi.advanceTimersByTimeAsync(300000);
    });
    expect(global.fetch).toHaveBeenCalledTimes(2);

    // Fast-forward another 5 minutes
    await act(async () => {
      await vi.advanceTimersByTimeAsync(300000);
    });
    expect(global.fetch).toHaveBeenCalledTimes(3);
  });
});