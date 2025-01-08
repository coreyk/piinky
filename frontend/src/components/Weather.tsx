import React, { useEffect, useState } from 'react';
import 'weather-icons-npm/css/weather-icons.css';
import { WeatherData, ForecastData } from '../types/weather';
import { cn, titleCase } from '../lib/utils';

const Weather: React.FC = () => {
  const API_HOST = import.meta.env.VITE_API_HOST || window.location.protocol + '//' + window.location.hostname;
  const API_PORT = import.meta.env.VITE_API_PORT || '8000';
  const API_URL = `${API_HOST}:${API_PORT}`;
  const [currentWeather, setCurrentWeather] = useState<WeatherData | null>(null);
  const [hourlyForecast, setHourlyForecast] = useState<ForecastData[]>([]);

  useEffect(() => {
    const fetchWeather = async () => {
      try {
        const response = await fetch(`${API_URL}/api/weather`);
        if (!response.ok) {
          throw new Error('Failed to fetch weather data');
        }

        const weatherData = await response.json();
        setCurrentWeather(weatherData);
        setHourlyForecast(weatherData.hourly_forecast);
      } catch (error) {
        console.error('Error fetching weather:', error);
      }
    };

    // Initial fetch
    let mounted = true;
    if (mounted) {
      fetchWeather();
    }

    const interval = setInterval(fetchWeather, 300000);

    return () => {
      mounted = false;
      clearInterval(interval);
    };
  }, []);

  if (!currentWeather) return null;

  return (
    <div className="flex flex-col items-center font-weather">
      <div className="flex items-center gap-0.5 text-black-900">
        <div className="flex flex-col items-center">
          <div className="flex items-center gap-0.5">
            <i className={`text-xl wi wi-owm-${currentWeather.icon}`}></i>
            <div className="text-xl font-bold">
              {Math.round(currentWeather.temperature.temp)}°
            </div>
            <div className="flex flex-row items-end gap-0.5 pl-1">
              <span className="text-red-500 font-bold">{Math.round(currentWeather.temperature.max)}°</span>/
              <span className="text-blue-500 font-bold">{Math.round(currentWeather.temperature.min)}°</span>
            </div>
            <div className="font-bold pl-1">
              {currentWeather.humidity}<i className="ml-0.25 wi wi-humidity" style={{ fontSize: '0.9rem' }}></i>
            </div>
            <div className="font-bold ml-0.5">
              <i className="ml-0.25 wi wi-strong-wind"></i> {Math.round(currentWeather.wind_speed)}<span style={{ fontSize: '0.7rem' }}>mph</span>
            </div>
          </div>
          <p className="text-center font-bold" style={{ fontSize: '0.6rem', marginTop: '-0.25rem' }}>{titleCase(currentWeather.detailed_status)}</p>
        </div>
        <div className="flex gap-0.5 border-l border-blue-800 pl-0.5 ml-0.75">
          {hourlyForecast.filter((_, index) => index % 2 === 0).slice(0, 6).map((item, index) => (
            <div key={index} className="flex flex-col items-center min-w-4rem">
              <div className="flex items-center pl-1">
                <div className="font-bold">
                  {new Date(Number(item.timestamp) * 1000).toLocaleTimeString([], { hour: 'numeric', minute: undefined, hour12: true }).toLowerCase().replace(/\s?([ap])m$/, '$1')}
                </div>
                <div className="font-bold">
                  <i className={cn("mr-0.5 ml-0.5 wi", `wi-owm-${item.icon}`)}></i>{Math.round(item.temperature.temp)}°
                </div>
              </div>
              <p className="text-center font-bold truncate max-w-16" style={{ fontSize: '0.6rem' }}>{item.status}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Weather;