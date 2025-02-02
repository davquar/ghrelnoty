package ghrelnoty

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"it.davquar/gitrelnoty/pkg/release"
)

type dummyRepoConfig struct {
	Type        string `yaml:"type"`
	Name        string `yaml:"name"`
	Destination string `yaml:"destination"`
}

type dummyReleaser struct {
	dummyRepoConfig
}

func (r dummyReleaser) GetLatestRelease(_ context.Context) (release.Release, RateLimitData, error) {
	return release.Release{
		Project:     "name",
		Author:      "author",
		Version:     "v1.2.3",
		Description: "some test description",
		URL:         "https://github.com/davquar/ghrelnoty/releases/tag/v1.2.3",
	}, RateLimitData{}, nil
}

func (r dummyReleaser) Config() RepositoryConfig {
	return RepositoryConfig(r.dummyRepoConfig)
}

type dummyNotifier struct{}

func (d dummyNotifier) Notify(release release.Release) error {
	// nolint (allow printing in the test output)
	fmt.Println("dummy notification", release)
	return nil
}

func TestWork(t *testing.T) {
	f, err := os.CreateTemp("", "ghrelnoty-")
	if err != nil {
		t.Fatalf("error creating temporary file: %v", err)
	}

	cfg := Config{
		CheckEvery:   1 * time.Second,
		SleepBetween: 1 * time.Second,
		DBPath:       f.Name(),
	}

	s, err := New(cfg)
	if err != nil {
		t.Fatalf("error creating service: %v", err)
	}

	s.Releasers = []Releaser{
		dummyReleaser{
			dummyRepoConfig{
				Name:        "author/name",
				Destination: "noop",
			},
		},
	}
	s.Notifiers = map[string]Notifier{
		"noop": dummyNotifier{},
	}

	c := make(chan error, 1)
	go s.Work(c)

	err = <-c
	if err != nil {
		t.Fatalf("%v", err)
	}
}
