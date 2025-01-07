import React from 'react';
import Event from './Event';
import { DayData } from '../types/calendar';
import { cn } from "@/lib/utils"
import 'weather-icons-npm/css/weather-icons.css';

interface CalendarDayProps {
  dayData: DayData;
  calendarData: any;
  dailyForecast?: Array<{
    timestamp: string;
    icon: string;
  }>;
}

const CalendarDay: React.FC<CalendarDayProps> = ({ dayData, calendarData, dailyForecast }) => {
  const dayNumber = new Date(dayData.date).getDate();
  const events = dayData.events || [];

  // Filter out recycling events from the regular event list
  const regularEvents = events.filter(event => event.colorId !== 'recycling');

  // Check for special recycling events
  const hasCardboard = events.some(event =>
    event.colorId === 'recycling' &&
    event.summary === 'Cardboard'
  );
  const hasCans = events.some(event =>
    event.colorId === 'recycling' &&
    event.summary === 'Cans' &&
    !event.start?.dateTime
  );
  const hasTrash = events.some(event =>
    event.colorId === 'recycling' &&
    event.summary === 'Trash' &&
    !event.start?.dateTime
  );

  // Find matching weather forecast for this day
  const weatherForDay = dailyForecast?.find(forecast => {
    const forecastDate = new Date(Number(forecast.timestamp) * 1000);
    return forecastDate.getDate() === dayNumber;
  });

  return (
    <div
      className={cn(
        'bg-white p-0.5 border-b border-r border-blue-500 relative',
        !dayData.isCurrentMonth && 'bg-slate-200'
      )}
      style={{
        minHeight: `${Math.floor(412 / calendarData.numberOfWeeks)}px`
      }}
    >
      <div className="flex justify-between items-start">
        <div className="flex flex-row items-start gap-0.5">
          <div className={cn(
            'px-0.5 w-7 h-7 rounded-full flex items-center justify-center',
            dayData.isToday && 'bg-blue-600 text-white',
            !dayData.isToday && 'bg-white border-2 border-blue-800 text-black-700',
            !dayData.isCurrentMonth && 'text-black-800 bg-red-200'
          )}>
            {dayNumber}
          </div>
          {weatherForDay && (
            <i className={`wi wi-owm-${weatherForDay.icon} ml-1 pt-1 text-sm`} style={{ marginTop: '-2px' }}></i>
          )}
        </div>
        {(hasCardboard || hasCans || hasTrash) && (
          <div className="flex flex-col gap-0.5 ml-2 text-base" style={{ fontFamily: "'Segoe UI Emoji', 'Apple Color Emoji', 'Noto Color Emoji', sans-serif" }}>
            {hasCardboard && <span className="emoji" title="Cardboard Recycling">📦♻️</span>}
            {hasCans && <span className="emoji" title="Can Recycling">🥤♻️</span>}
            {hasTrash && <span className="emoji" title="Trash Pickup">🗑️</span>}
          </div>
        )}
      </div>
      <div className="mt-1 space-y-1 max-h-[100px]">
        {regularEvents.map((event, index) => (
          <Event
            key={`${event.id}-${index}`}
            event={event}
            totalEvents={regularEvents.length}
          />
        ))}
        {regularEvents.length > 5 && (
          <div className="text-xs text-black-500 pl-2">
            +{regularEvents.length - 2} more
          </div>
        )}
      </div>
    </div>
  );
};

export default CalendarDay;