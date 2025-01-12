package ghrelnoty

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	dstsmtp "it.davquar/gitrelnoty/internal/ghrelnoty/destinations/smtp"
)

// Config holds the app's configuration.
type Config struct {
	LogLevel     slog.Level                   `yaml:"log_level"`
	DBPath       string                       `yaml:"db_path"`
	CheckEvery   time.Duration                `yaml:"check_every"`
	SleepBetween time.Duration                `yaml:"sleep_between"`
	Repositories []RepositoryConfig           `yaml:"repositories"`
	Destinations map[string]DestinationConfig `yaml:"destinations"`
	MetricsPort  int                          `yaml:"metrics_port"`
}

// RepositoryConfig holds data needed to identify the repository to watch
// and the destination to send notifications to.
type RepositoryConfig struct {
	Type        string `yaml:"type"`
	Name        string `yaml:"name"`
	Destination string `yaml:"destination"`
}

// DestinationConfig holds specific notification settings.
// Config is a different struct based on Type.
type DestinationConfig struct {
	Type   string `yaml:"type"`
	Config interface{}
}

// SeparateName returns a pair of repo-owner and repo-name, from a string
// like repo-owner/repo-name
func (r RepositoryConfig) SeparateName() (string, string) {
	repo := strings.Split(r.Name, "/")
	return repo[0], repo[1]
}

// UnmarshalYAML implements custom unmarshaling logic to produce the
// appropriate DestinationConfig.Config implementation based on DestinationConfig.Type.
func (dc *DestinationConfig) UnmarshalYAML(value *yaml.Node) error {
	var temp struct {
		Type   string    `yaml:"type"`
		Config yaml.Node `yaml:"config"`
	}

	if err := value.Decode(&temp); err != nil {
		return err
	}

	dc.Type = temp.Type

	switch dc.Type {
	case "smtp":
		var d dstsmtp.Destination
		if err := temp.Config.Decode(&d); err != nil {
			return err
		}
		dc.Config = &d
	default:
		return fmt.Errorf("unknown notifier type %s", dc.Type)
	}

	return nil
}

// Notifier returns the Notifier associated with the DestinationConfig.
func (dc *DestinationConfig) Notifier() (Notifier, error) {
	notifier, ok := dc.Config.(Notifier)
	if !ok {
		return nil, fmt.Errorf("invalid notifier configuration")
	}
	return notifier, nil
}
