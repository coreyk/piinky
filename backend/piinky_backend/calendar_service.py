"""
This module provides the CalendarService class for interacting with Google Calendar API,
loading configuration, and retrieving calendar events.
"""

from datetime import datetime, timedelta
import json
from fastapi import HTTPException
from google.oauth2.service_account import Credentials
from googleapiclient.discovery import build

class CalendarService:
    """
    This class handles interactions with the Google Calendar API,
    including loading configuration and retrieving calendar events.
    """

    def __init__(self, credentials_file, config_file):
        self.load_config(credentials_file, config_file)

    def load_config(self, credentials_file, config_file):
        """
        Load configuration from the specified config file and set up
        Google Calendar credentials from the provided credentials file.
        """
        with open(config_file, 'r', encoding='utf-8') as f:
            config = json.load(f)
        self.calendars = config["calendars"]
        self.start_on_sunday = config["start_on_sunday"]
        self.number_of_weeks = config["number_of_weeks"]
        self.credentials = Credentials.from_service_account_file(credentials_file)
        self.service = build("calendar", "v3", credentials=self.credentials)

    async def get_calendar_data(self):
        """
        Retrieve calendar data for the current week, including events
        from configured calendars.
        """
        try:
            today = datetime.now().astimezone()
            if self.start_on_sunday:
                week_start = today - timedelta(days=((today.weekday() + 1) % 7))
            else:
                week_start = today - timedelta(days=today.weekday())

            end_date = week_start + timedelta(weeks=self.number_of_weeks)

            events = await self.get_events(week_start.isoformat(), end_date.isoformat())
            return {
                "startDate": week_start.isoformat(),
                "endDate": end_date.isoformat(),
                "numberOfWeeks": self.number_of_weeks,
                "startOnSunday": self.start_on_sunday,
                "events": events
            }
        except Exception as e:
            raise HTTPException(status_code=500, detail=str(e)) from e

    async def get_events(self, start_date: str, end_date: str) -> list:
        """
        Retrieve events from configured calendars within the specified date range.

        Args:
            start_date (str): The start date in ISO format.
            end_date (str): The end date in ISO format.

        Returns:
            list: A list of events from the calendars.
        """
        try:
            all_events = []
            for calendar in self.calendars:
                calendar_id = calendar['calendar_id']
                events_result = self.service.events().list(
                    calendarId=calendar_id,
                    timeMin=start_date,
                    timeMax=end_date,
                    singleEvents=True,
                    orderBy='startTime'
                ).execute()

                # Add calendar color to each event
                events = events_result.get('items', [])
                for event in events:
                    event['colorId'] = calendar.get('color', '#000')
                all_events.extend(events)

            return all_events
        except Exception as e:
            raise HTTPException(status_code=500, detail=str(e)) from e
