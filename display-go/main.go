package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/coreyk/piinky/display-go/display"
)

func main() {
	displayService, err := display.NewDisplayService()
	if err != nil {
		log.Fatalf("Failed to initialize display service: %v", err)
	}

	screenshotPath := "piinky.png"

	for {
		// Create Chrome instance
		ctx, cancel := chromedp.NewContext(context.Background())

		// Create a timeout
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)

		time.Sleep(10 * time.Second)

		// Navigate to the page and capture screenshot
		var buf []byte
		if err := chromedp.Run(ctx,
			chromedp.EmulateViewport(800, 480),
			chromedp.Navigate("http://localhost:3000"),
			chromedp.Sleep(3*time.Second),
			chromedp.CaptureScreenshot(&buf),
		); err != nil {
			log.Printf("Failed to take screenshot: %v", err)
			cancel()
			continue
		}

		// Save the screenshot
		if err := os.WriteFile(screenshotPath, buf, 0644); err != nil {
			log.Printf("Failed to save screenshot: %v", err)
			cancel()
			continue
		}

		if err := displayService.UpdateDisplay(screenshotPath); err != nil {
			log.Printf("Failed to update display: %v", err)
		}

		cancel()
		time.Sleep(4 * time.Hour)
	}
}
