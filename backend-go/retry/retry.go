package retry

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxAttempts     int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	JitterFraction  float64 // 0.25 means ±25% jitter
}

// DefaultConfig returns sensible defaults for API calls
func DefaultConfig() Config {
	return Config{
		MaxAttempts:    5,
		BaseDelay:      time.Second,
		MaxDelay:       30 * time.Second,
		JitterFraction: 0.25,
	}
}

// calculateDelay returns exponential backoff delay with jitter
func (c Config) calculateDelay(attempt int) time.Duration {
	delay := c.BaseDelay * time.Duration(1<<uint(attempt))
	if delay > c.MaxDelay {
		delay = c.MaxDelay
	}
	// Add jitter
	jitter := float64(delay) * c.JitterFraction * (2*rand.Float64() - 1)
	return delay + time.Duration(jitter)
}

// IsTransientError determines if an error is likely transient and worth retrying
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network errors are generally transient
	if _, ok := err.(net.Error); ok {
		return true
	}

	// Common transient error patterns
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"timeout",
		"temporary failure",
		"too many open files",
		"network is unreachable",
		"i/o timeout",
		"EOF",
		"broken pipe",
	}

	for _, pattern := range transientPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}

// IsTransientStatusCode returns true for HTTP status codes that are worth retrying
func IsTransientStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,      // 429
		http.StatusInternalServerError,   // 500
		http.StatusBadGateway,            // 502
		http.StatusServiceUnavailable,    // 503
		http.StatusGatewayTimeout:        // 504
		return true
	}
	return false
}

// Do executes a function with retry logic
func Do(ctx context.Context, cfg Config, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry if error is not transient
		if !IsTransientError(err) {
			return err
		}

		// Don't wait after last attempt
		if attempt < cfg.MaxAttempts-1 {
			delay := cfg.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return fmt.Errorf("after %d attempts: %w", cfg.MaxAttempts, lastErr)
}

// DoWithResult executes a function that returns a result with retry logic
func DoWithResult[T any](ctx context.Context, cfg Config, operation func() (T, error)) (T, error) {
	var lastErr error
	var zero T

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		result, err := operation()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry if error is not transient
		if !IsTransientError(err) {
			return zero, err
		}

		// Don't wait after last attempt
		if attempt < cfg.MaxAttempts-1 {
			delay := cfg.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return zero, fmt.Errorf("after %d attempts: %w", cfg.MaxAttempts, lastErr)
}
