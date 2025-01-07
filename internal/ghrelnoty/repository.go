package ghrelnoty

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
	"it.davquar/gitrelnoty/internal/metrics"
)

type GitHubRepository struct {
	RepositoryConfig
}

func (r GitHubRepository) Config() RepositoryConfig {
	return r.RepositoryConfig
}

// GetLatestRelease gets the latest release name for the repository and the current rate limits.
func (r GitHubRepository) GetLatestRelease(ctx context.Context) (string, RateLimitData, error) {
	client := github.NewClient(nil)

	author, repo := r.SeparateName()
	release, resp, err := client.Repositories.GetLatestRelease(ctx, author, repo)

	rateLimitData, errr := makeRateLimitData(resp.Header)
	if errr != nil {
		return "", rateLimitData, fmt.Errorf("can't get rate limit data: %w", err)
	}

	metrics.SetRateLimitValue(float64(rateLimitData.Limit))
	metrics.SetRateLimitUsedValue(float64(rateLimitData.Used))

	rateLimitErr := isRateLimited(err)
	if rateLimitErr != nil {
		return "", rateLimitData, rateLimitErr
	}

	if err != nil {
		return "", rateLimitData, fmt.Errorf("%s: %w", r.Name, err)
	}

	return release.GetName(), rateLimitData, nil
}
