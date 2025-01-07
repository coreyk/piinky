"""
This module contains the main function to run the display service,
which updates the display periodically by taking screenshots.
"""

import asyncio
from playwright.async_api import async_playwright
from display_service import DisplayService

async def main():
    """Main function to run the display service and update the display periodically."""
    display_service = DisplayService()

    try:
        while True:
            await asyncio.sleep(5)

            async with async_playwright() as p:
                browser = await p.chromium.launch(headless=True)
                page = await browser.new_page(viewport={'width': 800, 'height': 480})

                await page.goto("http://localhost:3000", timeout=30000)

                # Wait for the page to load, weather is slow
                await asyncio.sleep(5)

                # Take screenshot and update display
                screenshot_path = "piinky.png"
                await page.screenshot(path=screenshot_path)
                await display_service.update_display(screenshot_path)

                await browser.close()

                await asyncio.sleep(14400)

    except Exception as e:
        print(f"Error updating display: {e}")
    finally:
        print("Shutting down gracefully...")

if __name__ == "__main__":
    asyncio.run(main())
