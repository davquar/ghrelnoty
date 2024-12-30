package ghrelnoty

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"it.davquar/gitrelnot/internal/store"
)

type Service struct {
	Config Config
	Store  store.Store
}

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
		return Service{}, fmt.Errorf("cannot open db: %w", err)
	}
	s.Store = db

	return s, nil
}

func (s Service) Work() {
	ticker := time.NewTicker(s.Config.CheckEvery)
	for ; true; <-ticker.C {
		for _, repo := range s.Config.Repositories {
			time.Sleep(s.Config.SleepBetween)
			ctx := context.Background()
			release, rateLimitData, err := repo.GetLatestRelease(ctx)
			if rateLimitData.IsAtRisk() {
				slog.WarnContext(ctx, "currently at risk of hitting rate limit. Pausing for 30 minutes")
				time.Sleep(30 * time.Minute)
			}

			if err != nil {
				slog.ErrorContext(ctx, "can't get latest release", slog.Any("err", err))

				var errRateLimited *RateLimitError
				if errors.As(err, &errRateLimited) {
					slog.ErrorContext(ctx, "hit rate limit: resuming activities at", slog.Any("time", rateLimitData.ResetAt))
					time.Sleep(time.Until(rateLimitData.ResetAt))
				}
				continue
			}

			changed, err := s.Store.CompareAndSet(repo.Name, release)
			if err != nil {
				slog.ErrorContext(ctx, "can't store in db", slog.String("repo", repo.Name), slog.Any("err", err))
			}

			slog.Debug("got data", slog.String("repo", repo.Name), slog.String("release", release), slog.Bool("changed", changed))

			if changed {
				dst, ok := s.Config.Destinations[repo.Destination]
				if !ok {
					slog.Error("destination not found", slog.String("destination", repo.Destination))
					continue
				}

				err := dst.Notify(repo.Name, release)
				if err != nil {
					slog.Error("cannot notify", slog.Any("err", err))
					continue
				}
			}
		}
	}
}

func (s *Service) Close() {
	s.Store.Close()
}
