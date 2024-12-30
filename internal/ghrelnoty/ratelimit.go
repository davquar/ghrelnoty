package ghrelnoty

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v68/github"
)

type RateLimitError struct {
	Type string `json:"type"`
}

func (e RateLimitError) Error() string {
	return fmt.Sprintf("Rate limited: %s", e.Type)
}

type RateLimitData struct {
	Limit     int
	Remaining int
	Used      int
	ResetAt   time.Time
}

func (r RateLimitData) GetUsedPercent() float64 {
	return float64(r.Used) / float64(r.Limit)
}

func (r RateLimitData) IsAtRisk() bool {
	return r.GetUsedPercent() > 0.8
}

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
