package ghrelnoty

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-github/v68/github"
)

func TestNotRateLimited(t *testing.T) {
	h := http.Header{}
	h.Set("x-ratelimit-limit", "60")
	h.Set("x-ratelimit-remaining", "60")
	h.Set("x-ratelimit-used", "0")
	h.Set("x-ratelimit-reset", "1735577226")

	d, err := makeRateLimitData(h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.Limit != 60 || d.Remaining != 60 || d.Used != 0 {
		t.Fatalf("expected [limit, remainig, used] of [60, 60, 0], got [%d, %d, %d]", d.Limit, d.Remaining, d.Used)
	}

	if d.IsAtRisk() {
		t.Fatal("expected no risk; got risk")
	}

	usedPercent := d.GetUsedPercent()
	if usedPercent != 0 {
		t.Fatalf("expected used 0, got %f", usedPercent)
	}
}

func TestRisk(t *testing.T) {
	h := http.Header{}
	h.Set("x-ratelimit-limit", "100")
	h.Set("x-ratelimit-remaining", "10")
	h.Set("x-ratelimit-used", "90")
	h.Set("x-ratelimit-reset", "1735577226")

	d, err := makeRateLimitData(h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !d.IsAtRisk() {
		t.Fatal("expected risk; got no risk")
	}

	usedPercent := d.GetUsedPercent()
	if usedPercent != 0.9 {
		t.Fatalf("expected used 0.9, got %f", usedPercent)
	}
}

func TestRateLimited(t *testing.T) {
	gitHubRateLimited := &github.RateLimitError{}

	err := isRateLimited(gitHubRateLimited)
	if err == nil {
		t.Fatal("expected an actual rate limit error, got nil")
	}

	var errRateLimited *RateLimitError
	if !errors.As(err, &errRateLimited) {
		t.Fatalf("expected type RateLimitError, got other: %v", err)
	}

	if errRateLimited.Type != "primary" {
		t.Fatalf("expected type primary, got %s", errRateLimited.Type)
	}
}

func TestRateLimitedAbuse(t *testing.T) {
	gitHubRateLimited := &github.AbuseRateLimitError{}

	err := isRateLimited(gitHubRateLimited)
	if err == nil {
		t.Fatal("expected an actual rate limit error, got nil")
	}

	var errRateLimited *RateLimitError
	if !errors.As(err, &errRateLimited) {
		t.Fatalf("expected type RateLimitError, got other: %v", err)
	}

	if errRateLimited.Type != "secondary" {
		t.Fatalf("expected type secondary, got %s", errRateLimited.Type)
	}
}
