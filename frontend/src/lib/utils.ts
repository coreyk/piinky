import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export const formatTime = (timeString: string | undefined): string => {
  if (!timeString) return '';
  const time = new Date(timeString);
  return time.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: time.getMinutes() === 0 ? undefined : '2-digit',
    hour12: true
  }).toLowerCase().replace(/\s?([ap])m$/, '$1');
};

export const titleCase = (str: string): string => {
  return str.replace(/_/g, ' ').replace(/\b\w/g, (char) => char.toUpperCase());
};
