package ghrelnoty

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type Service struct {
	Config Config
}

func New(config Config) Service {
	s := Service{
		Config: config,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.LogLevel,
	}))
	slog.SetDefault(logger)

	return s
}

func (s Service) Work() error {
	ticker := time.NewTicker(s.Config.CheckEvery)
	go func() {
		for range ticker.C {
			for _, repo := range s.Config.Repositories {
				ctx := context.Background()
				err := repo.GetReleases(ctx, false)
				if err != nil {
					slog.ErrorContext(ctx, "can't get releases: %w", slog.Any("err", err))
				}
			}
		}
	}()
	return nil
}
