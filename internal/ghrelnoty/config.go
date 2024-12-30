package ghrelnoty

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
)

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

func (r Repository) GetLatestRelease(ctx context.Context) (string, error) {
	client := github.NewClient(nil)

	author, repo := r.SeparateName()
	release, _, err := client.Repositories.GetLatestRelease(ctx, author, repo)
	if err != nil {
		return "", fmt.Errorf("%s: %w", r.Name, err)
	}

	return release.GetName(), nil
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
