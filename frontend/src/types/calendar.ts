export interface CalendarData {
  startDate: string;
  endDate: string;
  startOnSunday: boolean;
  numberOfWeeks: number;
  events: DayData[];
}

export interface DayData {
  date: string;
  events: CalendarEvent[];
  isToday: boolean;
  isCurrentMonth: boolean;
}

export interface CalendarEvent {
  id: string;
  summary: string;
  description?: string;
  location?: string;
  start: {
    date?: string;
    dateTime?: string;
    timeZone: string;
  };
  end: {
    date?: string;
    dateTime?: string;
    timeZone: string;
  };
  colorId?: string;
}