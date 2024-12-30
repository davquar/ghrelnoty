package ghrelnoty

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v68/github"
)

// RateLimitError is a generic error type that represents a rate limiting error
// of a specific type: "primary", or "secondary" if it is classified as abuse.
type RateLimitError struct {
	Type string `json:"type"`
}

func (e RateLimitError) Error() string {
	return fmt.Sprintf("Rate limited: %s", e.Type)
}

// RateLimitData holds counters that describe the current usage of the API, wrt
// GitHub's rate limits. This data is extracted from GitHub's HTTP responses.
type RateLimitData struct {
	Limit     int
	Remaining int
	Used      int
	ResetAt   time.Time
}

// GetUsedPercent returns the current usage percentage, compared to the limits.
// It is expressed as a float in the iterval 0.0 and 1.0.
func (r RateLimitData) GetUsedPercent() float64 {
	return float64(r.Used) / float64(r.Limit)
}

// IsAtRisk returns true if there is a risk of hitting the rate limit, namely when
// more than 80% of the limits are used.
func (r RateLimitData) IsAtRisk() bool {
	return r.GetUsedPercent() > 0.8
}

// makeRateLimitData returns the RateLimitData after extrating needed values from
// the given HTTP headers.
func makeRateLimitData(headers http.Header) (RateLimitData, error) {
	limitStr := headers.Get("x-ratelimit-limit")
	remainingStr := headers.Get("x-ratelimit-remaining")
	usedStr := headers.Get("x-ratelimit-used")
	resetStr := headers.Get("x-ratelimit-reset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return RateLimitData{}, fmt.Errorf("convert limit: %w", err)
	}
	remaining, err := strconv.Atoi(remainingStr)
	if err != nil {
		return RateLimitData{}, fmt.Errorf("convert remaining: %w", err)
	}
	used, err := strconv.Atoi(usedStr)
	if err != nil {
		return RateLimitData{}, fmt.Errorf("convert used: %w", err)
	}
	reset, err := strconv.ParseInt(resetStr, 10, 64)
	if err != nil {
		return RateLimitData{}, fmt.Errorf("parse reset: %w", err)
	}
	resetTime := time.Unix(reset, 0)

	return RateLimitData{
		Limit:     limit,
		Remaining: remaining,
		Used:      used,
		ResetAt:   resetTime,
	}, nil
}

// isRateLimited returns an RateLimitError if the given error is a GitHub
// rate limiting error, namely a github.RateLimitError or github.AbuseRateLimitError.
func isRateLimited(err error) error {
	var rateLimitError *github.RateLimitError
	var abuseRateLimitError *github.AbuseRateLimitError

	if errors.As(err, &rateLimitError) {
		return &RateLimitError{
			Type: "primary",
		}
	}

	if errors.As(err, &abuseRateLimitError) {
		return &RateLimitError{
			Type: "secondary",
		}
	}

	return nil
}
