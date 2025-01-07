package ghrelnoty

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

type dummyRepoConfig struct {
	Type        string `yaml:"type"`
	Name        string `yaml:"name"`
	Destination string `yaml:"destination"`
}

type dummyReleaser struct {
	dummyRepoConfig
}

func (r dummyReleaser) GetLatestRelease(_ context.Context) (string, RateLimitData, error) {
	return "v0.1.2", RateLimitData{}, nil
}

func (r dummyReleaser) Config() RepositoryConfig {
	return RepositoryConfig(r.dummyRepoConfig)
}

type dummyNotifier struct{}

func (d dummyNotifier) Notify(owner string, repo string) error {
	// nolint (allow printing in the test output)
	fmt.Println("dummy notification", owner, repo)
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
