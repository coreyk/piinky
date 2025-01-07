import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import CalendarDay from './CalendarDay';

describe('CalendarDay', () => {
  const mockCalendarData = {
    numberOfWeeks: 4
  };

  // Create a date at midnight UTC
  const testDate = new Date('2024-01-15T00:00:00.000Z');
  const mockDayData = {
    date: testDate.toISOString(),
    isCurrentMonth: true,
    isToday: false,
    events: []
  };

  it('renders day number correctly', () => {
    const { container } = render(
      <CalendarDay
        dayData={mockDayData}
        calendarData={mockCalendarData}
      />
    );

    // Get the actual rendered number
    const dayNumber = container.querySelector('.rounded-full')?.textContent?.trim();
    expect(dayNumber).toBe('15');
  });

  it('displays weather icon when forecast is provided', () => {
    const mockForecast = [{
      timestamp: Math.floor(testDate.getTime() / 1000).toString(),
      icon: '800' // Clear sky
    }];

    const { container } = render(
      <CalendarDay
        dayData={mockDayData}
        calendarData={mockCalendarData}
        dailyForecast={mockForecast}
      />
    );

    expect(container.querySelector('.wi-owm-800')).toBeInTheDocument();
  });

  it('does not display weather icon when no matching forecast', () => {
    const nextDay = new Date(testDate);
    nextDay.setDate(nextDay.getDate() + 1);

    const mockForecast = [{
      timestamp: Math.floor(nextDay.getTime() / 1000).toString(),
      icon: '800'
    }];

    const { container } = render(
      <CalendarDay
        dayData={mockDayData}
        calendarData={mockCalendarData}
        dailyForecast={mockForecast}
      />
    );

    expect(container.querySelector('.wi-owm-800')).not.toBeInTheDocument();
  });

  it('displays both weather and recycling icons when both are present', () => {
    const mockForecast = [{
      timestamp: Math.floor(testDate.getTime() / 1000).toString(),
      icon: '800'
    }];

    const dayDataWithRecycling = {
      ...mockDayData,
      events: [{
        id: '1',
        summary: 'Cardboard',
        colorId: 'recycling',
        start: { date: testDate.toISOString().split('T')[0], timeZone: 'UTC' },
        end: { date: testDate.toISOString().split('T')[0], timeZone: 'UTC' }
      }]
    };

    const { container } = render(
      <CalendarDay
        dayData={dayDataWithRecycling}
        calendarData={mockCalendarData}
        dailyForecast={mockForecast}
      />
    );

    expect(container.querySelector('.wi-owm-800')).toBeInTheDocument();
    expect(screen.getByTitle('Cardboard Recycling')).toBeInTheDocument();
  });

  it('handles different weather icon codes correctly', () => {
    const mockForecast = [{
      timestamp: Math.floor(testDate.getTime() / 1000).toString(),
      icon: '500' // Rain
    }];

    const { container } = render(
      <CalendarDay
        dayData={mockDayData}
        calendarData={mockCalendarData}
        dailyForecast={mockForecast}
      />
    );

    expect(container.querySelector('.wi-owm-500')).toBeInTheDocument();
  });
});