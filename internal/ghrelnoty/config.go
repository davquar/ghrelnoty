package ghrelnoty

import (
	"log/slog"
	"time"
)

// Config holds the app's configuration.
type Config struct {
	LogLevel     slog.Level             `yaml:"log_level"`
	DBPath       string                 `yaml:"db_path"`
	CheckEvery   time.Duration          `yaml:"check_every"`
	SleepBetween time.Duration          `yaml:"sleep_between"`
	Repositories []Repository           `yaml:"repositories"`
	Destinations map[string]Destination `yaml:"destinations"`
}
