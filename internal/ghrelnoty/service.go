package ghrelnoty

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"it.davquar/gitrelnoty/internal/metrics"
	"it.davquar/gitrelnoty/internal/store"
)

// Service holds the app's configuration and an instance of the KV store
// in which release data is saved for each repository.
type Service struct {
	Config Config
	Store  store.Store
}

// New initializes logging, opens the database and returns a new Service.
func New(config Config) (Service, error) {
	s := Service{
		Config: config,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.LogLevel,
	}))
	slog.SetDefault(logger)

	db, err := store.Open(s.Config.DBPath)
	if err != nil {
		metrics.DBOpenError()
		return Service{}, fmt.Errorf("cannot open db: %w", err)
	}
	s.Store = db

	return s, nil
}

// Work is a coordinator function that calls functions to:
// - Check releases for the configured repositories.
// - Handle rate limiting prevention and remediation.
// - Write to the database.
// - Notify in case of a new release.
func (s Service) Work() {
	ticker := time.NewTicker(s.Config.CheckEvery)
	for ; true; <-ticker.C {
		for _, repo := range s.Config.Repositories {
			time.Sleep(s.Config.SleepBetween)
			ctx := context.Background()
			release, rateLimitData, err := repo.GetLatestRelease(ctx)

			if rateLimitData.IsAtRisk() {
				metrics.RateLimitRisk()
				slog.WarnContext(ctx, "currently at risk of hitting rate limit. Pausing for 30 minutes")
				time.Sleep(30 * time.Minute)
			}

			if err != nil {
				metrics.CannotGetRelease()
				slog.ErrorContext(ctx, "can't get latest release", slog.Any("err", err))

				var errRateLimited *RateLimitError
				if errors.As(err, &errRateLimited) {
					metrics.RateLimited()
					slog.ErrorContext(ctx, "hit rate limit: resuming activities at", slog.Any("time", rateLimitData.ResetAt))
					time.Sleep(time.Until(rateLimitData.ResetAt))
				}
				continue
			}

			changed, err := s.Store.CompareAndSet(repo.Name, release)
			if err != nil {
				metrics.DBError()
				slog.ErrorContext(ctx, "can't store in db", slog.String("repo", repo.Name), slog.Any("err", err))
			}

			slog.Debug("got data", slog.String("repo", repo.Name), slog.String("release", release), slog.Bool("changed", changed))

			if changed {
				metrics.NewReleaseFound()
				dst, ok := s.Config.Destinations[repo.Destination]
				if !ok {
					metrics.NotificationError()
					slog.Error("destination not found", slog.String("destination", repo.Destination))
					continue
				}

				err := dst.Notify(repo.Name, release)
				if err != nil {
					metrics.NotificationError()
					slog.Error("cannot notify", slog.Any("err", err))
					continue
				}
			}
		}
	}
}

// Close closes the Service's handles, currently only the database.
func (s *Service) Close() {
	s.Store.Close()
}
