package ghrelnoty

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	smtpd "it.davquar/gitrelnoty/internal/ghrelnoty/destinations/smtp"
	"it.davquar/gitrelnoty/internal/metrics"
	"it.davquar/gitrelnoty/internal/store"
	"it.davquar/gitrelnoty/pkg/release"
)

// Service holds the app's configuration and an instance of the KV store
// in which release data is saved for each repository.
type Service struct {
	Config    Config
	Releasers []Releaser
	Notifiers map[string]Notifier
	Store     store.Store
}

// Notifier is implemented by notification system (Destination)
type Notifier interface {
	Notify(release release.Release) error
}

type Releaser interface {
	GetLatestRelease(context.Context) (release.Release, RateLimitData, error)
	Config() RepositoryConfig
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

	err = s.initReleasers()
	if err != nil {
		return Service{}, fmt.Errorf("init releasers: %w", err)
	}

	err = s.initNotifiers()
	if err != nil {
		return Service{}, fmt.Errorf("init notifiers: %w", err)
	}

	return s, nil
}

func (s *Service) initReleasers() error {
	s.Releasers = make([]Releaser, 0, len(s.Config.Repositories))
	for _, repo := range s.Config.Repositories {
		switch repo.Type {
		case "github":
			s.Releasers = append(s.Releasers, GitHubRepository{repo})
		default:
			return fmt.Errorf("unknown repo type for %s", repo.Name)
		}
	}
	return nil
}

func (s *Service) initNotifiers() error {
	s.Notifiers = make(map[string]Notifier)
	for name, dst := range s.Config.Destinations {
		switch dst.Type {
		case "smtp":
			dstcfg, ok := dst.Config.(*smtpd.Destination)
			if !ok {
				return fmt.Errorf("assert %s of type smtp", name)
			}
			s.Notifiers[name] = dstcfg
		default:
			return fmt.Errorf("unknown type for %s", name)
		}
	}
	return nil
}

// WorkLoop calls Work on a regular time intervals.
func (s Service) WorkLoop() {
	ticker := time.NewTicker(s.Config.CheckEvery)
	for ; true; <-ticker.C {
		c := make(chan (error))
		go s.Work(c)
	}
}

// Work is a coordinator function that calls functions to:
// - Check releases for the configured repositories.
// - Handle rate limiting prevention and remediation.
// - Write to the database.
// - Notify in case of a new release.
func (s Service) Work(c chan (error)) {
	defer close(c)
	for _, repo := range s.Releasers {
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

		changed, err := s.Store.CompareAndSet(repo.Config().Name, release.Version)
		if err != nil {
			metrics.DBError()
			slog.ErrorContext(ctx, "can't store in db", slog.String("repo", repo.Config().Name), slog.Any("err", err))
			c <- err
		}

		slog.Debug("got data", slog.String("repo", repo.Config().Name), slog.String("release", release.Version), slog.Bool("changed", changed))

		if changed {
			metrics.NewReleaseFound()
			notifier, ok := s.Notifiers[repo.Config().Destination]
			if !ok {
				metrics.NotificationError()
				slog.Error("notifier not found", slog.String("destination", repo.Config().Destination))
				c <- errors.New("notifier not found")
				continue
			}

			err = notifier.Notify(release)
			if err != nil {
				metrics.NotificationError()
				slog.Error("cannot notify", slog.Any("err", err))
				c <- err
				continue
			}
		}
	}
}

// Close closes the Service's handles, currently only the database.
func (s *Service) Close() {
	s.Store.Close()
}
