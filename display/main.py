"""
This module contains the main function to run the display service,
which updates the display periodically by taking screenshots.
"""

import asyncio
import random
from datetime import datetime
from playwright.async_api import async_playwright, Error as PlaywrightError
from display_service import DisplayService

# Configuration
UPDATE_INTERVAL_SECONDS = 14400  # 4 hours
INITIAL_DELAY_SECONDS = 5
PAGE_LOAD_TIMEOUT_MS = 30000
PAGE_SETTLE_SECONDS = 5

# Quiet hours configuration (no updates during this window)
QUIET_HOURS_START = 22  # 10 PM
QUIET_HOURS_END = 6     # 6 AM

# Retry configuration
MAX_RETRIES = 5
BASE_RETRY_DELAY_SECONDS = 30
MAX_RETRY_DELAY_SECONDS = 600  # 10 minutes


def is_quiet_hours() -> bool:
    """Check if current time is within quiet hours (no display updates)."""
    current_hour = datetime.now().hour
    if QUIET_HOURS_START > QUIET_HOURS_END:
        # Quiet hours span midnight (e.g., 22:00 to 06:00)
        return current_hour >= QUIET_HOURS_START or current_hour < QUIET_HOURS_END
    else:
        # Quiet hours within same day
        return QUIET_HOURS_START <= current_hour < QUIET_HOURS_END


def seconds_until_quiet_hours_end() -> int:
    """Calculate seconds until quiet hours end."""
    now = datetime.now()
    current_hour = now.hour

    if current_hour >= QUIET_HOURS_START:
        # After start time, end is tomorrow
        hours_until_end = (24 - current_hour) + QUIET_HOURS_END
    else:
        # Before end time today
        hours_until_end = QUIET_HOURS_END - current_hour

    # Subtract current minutes/seconds for more accuracy
    seconds = hours_until_end * 3600 - now.minute * 60 - now.second
    return max(seconds, 60)  # At least 60 seconds


def get_retry_delay(attempt: int) -> float:
    """Calculate exponential backoff delay with jitter."""
    delay = min(BASE_RETRY_DELAY_SECONDS * (2 ** attempt), MAX_RETRY_DELAY_SECONDS)
    # Add jitter (±25%)
    jitter = delay * 0.25 * (2 * random.random() - 1)
    return delay + jitter


async def take_screenshot_with_retry(screenshot_path: str) -> bool:
    """
    Attempt to take a screenshot with retry logic.
    Returns True if successful, False if all retries failed.
    """
    for attempt in range(MAX_RETRIES):
        try:
            async with async_playwright() as p:
                browser = await p.chromium.launch(headless=True)
                try:
                    page = await browser.new_page(viewport={'width': 800, 'height': 480})
                    await page.goto("http://localhost:3000", timeout=PAGE_LOAD_TIMEOUT_MS)

                    # Wait for the page to load, weather is slow
                    await asyncio.sleep(PAGE_SETTLE_SECONDS)

                    # Take screenshot
                    await page.screenshot(path=screenshot_path)
                    return True
                finally:
                    await browser.close()

        except PlaywrightError as e:
            retry_delay = get_retry_delay(attempt)
            print(f"Screenshot attempt {attempt + 1}/{MAX_RETRIES} failed: {e}")
            if attempt < MAX_RETRIES - 1:
                print(f"Retrying in {retry_delay:.1f} seconds...")
                await asyncio.sleep(retry_delay)
            else:
                print("All screenshot attempts failed")
                return False

        except Exception as e:
            retry_delay = get_retry_delay(attempt)
            print(f"Unexpected error on attempt {attempt + 1}/{MAX_RETRIES}: {e}")
            if attempt < MAX_RETRIES - 1:
                print(f"Retrying in {retry_delay:.1f} seconds...")
                await asyncio.sleep(retry_delay)
            else:
                print("All screenshot attempts failed")
                return False

    return False


async def main():
    """Main function to run the display service and update the display periodically."""
    display_service = DisplayService()
    screenshot_path = "piinky.png"
    consecutive_failures = 0
    max_consecutive_failures = 10

    print("Display service starting...")

    while True:
        try:
            # Initial delay before first update
            await asyncio.sleep(INITIAL_DELAY_SECONDS)

            # Skip updates during quiet hours
            if is_quiet_hours():
                sleep_time = seconds_until_quiet_hours_end()
                print(f"Quiet hours active, sleeping until {QUIET_HOURS_END}:00 ({sleep_time}s)")
                await asyncio.sleep(sleep_time)
                continue

            # Take screenshot with retry logic
            if await take_screenshot_with_retry(screenshot_path):
                # Update display
                try:
                    await display_service.update_display(screenshot_path)
                    consecutive_failures = 0
                    print("Display updated successfully")
                except Exception as e:
                    print(f"Failed to update display hardware: {e}")
                    consecutive_failures += 1
            else:
                consecutive_failures += 1
                print(f"Failed to capture screenshot (consecutive failures: {consecutive_failures})")

            # If we've had too many consecutive failures, wait longer before next attempt
            if consecutive_failures >= max_consecutive_failures:
                extended_wait = UPDATE_INTERVAL_SECONDS // 2
                print(f"Too many consecutive failures ({consecutive_failures}), waiting {extended_wait}s before next attempt")
                await asyncio.sleep(extended_wait)
                consecutive_failures = 0  # Reset to try again
            else:
                # Normal update interval
                await asyncio.sleep(UPDATE_INTERVAL_SECONDS)

        except asyncio.CancelledError:
            print("Display service cancelled, shutting down...")
            break
        except Exception as e:
            # Catch-all for unexpected errors - log but don't exit
            print(f"Unexpected error in main loop: {e}")
            consecutive_failures += 1
            # Wait before retrying to avoid tight error loops
            await asyncio.sleep(BASE_RETRY_DELAY_SECONDS)

    print("Display service stopped")

if __name__ == "__main__":
    asyncio.run(main())
