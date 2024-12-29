package ghrelnoty

import (
	"context"
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

func (s Service) Work() error {
	ticker := time.NewTicker(s.Config.CheckEvery)
	go func() {
		for range ticker.C {
			for _, repo := range s.Config.Repositories {
				ctx := context.Background()
				release, err := repo.GetLatestRelease(ctx)
				if err != nil {
					slog.ErrorContext(ctx, "can't get latest release", slog.Any("err", err))
					continue
				}

				err = s.Store.Set(repo.Name, release)
				if err != nil {
					slog.ErrorContext(ctx, "can't store in db", slog.String("repo", repo.Name), slog.Any("err", err))
				}

				slog.Debug("updated", slog.String("repo", repo.Name), slog.String("release", release))
			}
		}
	}()
	return nil
}

func (s *Service) Close() {
	s.Store.Close()
}
