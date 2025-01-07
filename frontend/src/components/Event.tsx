import React from 'react';
import { CalendarEvent } from '../types/calendar';
import { bgColorMap, borderColorMap, textColorMap } from '../lib/colormap';
import { cn, formatTime } from "@/lib/utils"

interface EventProps {
  event: CalendarEvent;
  totalEvents: number;
}

const Event: React.FC<EventProps> = ({ event, totalEvents }) => {

  const isAllDay = !event.start?.dateTime;
  const eventBgColor = bgColorMap[event.colorId as keyof typeof bgColorMap] || bgColorMap['black'];
  const eventBorderColor = borderColorMap[event.colorId as keyof typeof borderColorMap] || borderColorMap['black'];
  const isCompactView = totalEvents > 3;

  return (
    <div className={cn(
      'text-xs font-semibold px-0.5 py-0.5 border-2 border-opacity-100',
      isAllDay && `rounded-sm ${eventBgColor} ${eventBorderColor}`,
      !isAllDay && `rounded-md ${eventBorderColor}`,
      isCompactView && 'truncate'
    )}>
      {!isAllDay && (
        <span className="font-semibold mr-1">
          {formatTime(event.start.dateTime)}
        </span>
      )}
      <span className={`${isAllDay ? 'text-white' : textColorMap[event.colorId as keyof typeof textColorMap] || textColorMap['black']}`}>
        {event.summary}
      </span>
    </div>
  );
};

export default Event;
