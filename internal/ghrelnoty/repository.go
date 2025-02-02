package ghrelnoty

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
	"it.davquar/gitrelnoty/internal/metrics"
	"it.davquar/gitrelnoty/pkg/release"
)

type GitHubRepository struct {
	RepositoryConfig
}

func (r GitHubRepository) Config() RepositoryConfig {
	return r.RepositoryConfig
}

// GetLatestRelease gets the latest Release for the repository and the current rate limits.
func (r GitHubRepository) GetLatestRelease(ctx context.Context) (release.Release, RateLimitData, error) {
	client := github.NewClient(nil)

	author, repo := r.SeparateName()
	repoRelease, resp, err := client.Repositories.GetLatestRelease(ctx, author, repo)

	rateLimitData, errr := makeRateLimitData(resp.Header)
	if errr != nil {
		return release.Release{}, rateLimitData, fmt.Errorf("can't get rate limit data: %w", err)
	}

	metrics.SetRateLimitValue(float64(rateLimitData.Limit))
	metrics.SetRateLimitUsedValue(float64(rateLimitData.Used))

	rateLimitErr := isRateLimited(err)
	if rateLimitErr != nil {
		return release.Release{}, rateLimitData, rateLimitErr
	}

	if err != nil {
		return release.Release{}, rateLimitData, fmt.Errorf("%s: %w", r.Name, err)
	}

	release := release.Release{
		Project:     repo,
		Author:      author,
		Version:     repoRelease.GetName(),
		Description: repoRelease.GetBody(),
		URL:         repoRelease.GetURL(),
	}
	return release, rateLimitData, nil
}
