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
	CheckEvery   time.Duration          `yaml:"check_every"`
	Repositories []Repository           `yaml:"repositories"`
	Destinations map[string]Destination `yaml:"destinations"`
}

type Repository struct {
	Name        string `yaml:"name"`
	Destination string `yaml:"destination"`
	Prereleases bool   `yaml:"prereleases"`
}

type Destination struct{}

func (r Repository) SeparateName() (string, string) {
	repo := strings.Split(r.Name, "/")
	return repo[0], repo[1]
}

func (r Repository) GetReleases(ctx context.Context, prereleases bool) error {
	client := github.NewClient(nil)

	author, repo := r.SeparateName()
	releases, _, err := client.Repositories.ListReleases(ctx, author, repo, &github.ListOptions{})
	if err != nil {
		return fmt.Errorf("cannot get last release for %s: %w", r.Name, err)
	}

	for _, release := range releases {
		if release.GetPrerelease() && !prereleases {
			continue
		}
		slog.Debug("found", slog.String("repo", r.Name), slog.String("release", release.GetName()))
	}

	return nil
}
