package ghrelnoty

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v68/github"
)

// Repository holds data needed to identify the repository to watch
// and the destination to send notifications to.
type Repository struct {
	Name        string `yaml:"name"`
	Destination string `yaml:"destination"`
}

// SeparateName returns a pair of repo-owner and repo-name, from a string
// like repo-owner/repo-name
func (r Repository) SeparateName() (string, string) {
	repo := strings.Split(r.Name, "/")
	return repo[0], repo[1]
}

// GetLatestRelease gets the latest release name for the repository and the current rate limits.
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
