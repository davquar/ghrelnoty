package ghrelnoty

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
)

type Config struct {
	LogLevel     slog.Level             `yaml:"log_level"`
	DBPath       string                 `yaml:"db_path"`
	CheckEvery   time.Duration          `yaml:"check_every"`
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
	Server   string `yaml:"server"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (d Destination) Notify() {
	fmt.Println("fake notification", d)
}
