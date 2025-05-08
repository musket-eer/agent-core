package utils

import (
	"errors"
	"strings"
	"time"
)

// CallWithBackoff retries the given LLM call on 429 errors with exponential backoff.
// llmCall should be a function that returns (string, error).
func CallWithBackoff(llmCall func() (string, error), maxAttempts int, initialDelay time.Duration) (string, error) {
	delay := initialDelay
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := llmCall()
		if err == nil {
			return resp, nil
		}
		// Check if error is 429 Too Many Requests
		errMsg := err.Error()
		if strings.Contains(errMsg, "429") || strings.Contains(strings.ToLower(errMsg), "too many requests") {
			lastErr = err
			time.Sleep(delay)
			delay *= 2
			continue
		}
		// For other errors, return immediately
		return "", err
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", errors.New("max retry attempts reached")
}
