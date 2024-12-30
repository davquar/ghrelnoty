package ghrelnoty

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
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

type Config struct {
	LogLevel     slog.Level             `yaml:"log_level"`
	DBPath       string                 `yaml:"db_path"`
	CheckEvery   time.Duration          `yaml:"check_every"`
	SleepBetween time.Duration          `yaml:"sleep_between"`
	Repositories []Repository           `yaml:"repositories"`
	Destinations map[string]Destination `yaml:"destinations"`
}

type Repository struct {
	Name        string `yaml:"name"`
	Destination string `yaml:"destination"`
}

func (r Repository) SeparateName() (string, string) {
	repo := strings.Split(r.Name, "/")
	return repo[0], repo[1]
}

func (r Repository) GetLatestRelease(ctx context.Context) (string, RateLimitData, error) {
	client := github.NewClient(nil)

	author, repo := r.SeparateName()
	release, resp, err := client.Repositories.GetLatestRelease(ctx, author, repo)

	rateLimitData, errr := makeRateLimitData(resp.Header)
	if errr != nil {
		return "", rateLimitData, fmt.Errorf("can't get rate limit data: %w", err)
	}

	rateLimitErr := isRateLimited(err)
	if rateLimitErr != nil {
		return "", rateLimitData, rateLimitErr
	}

	if err != nil {
		return "", rateLimitData, fmt.Errorf("%s: %w", r.Name, err)
	}

	return release.GetName(), rateLimitData, nil
}

func makeRateLimitData(headers http.Header) (RateLimitData, error) {
	limitStr := headers.Get("x-ratelimit-limit")
	remainingStr := headers.Get("x-ratelimit-remaining")
	usedStr := headers.Get("x-ratelimit-used")
	resetStr := headers.Get("x-ratelimit-reset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return RateLimitData{}, fmt.Errorf("convert: %w", err)
	}
	remaining, err := strconv.Atoi(remainingStr)
	if err != nil {
		return RateLimitData{}, fmt.Errorf("convert: %w", err)
	}
	used, err := strconv.Atoi(usedStr)
	if err != nil {
		return RateLimitData{}, fmt.Errorf("convert: %w", err)
	}
	reset, err := strconv.ParseInt(resetStr, 10, 64)
	if err != nil {
		return RateLimitData{}, fmt.Errorf("parse: %w", err)
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

type Destination struct {
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (d Destination) Notify(repo string, release string) error {
	subject := fmt.Sprintf("New release: %s:%s", repo, release)
	body := fmt.Sprintf("GHRelNoty detected a new release: %s:%s", repo, release)
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", d.From, d.To, subject, body))
	auth := smtp.PlainAuth("", d.From, d.Password, d.Host)
	err := smtp.SendMail(d.Host+":"+d.Port, auth, d.From, []string{d.To}, msg)

	return err
}
