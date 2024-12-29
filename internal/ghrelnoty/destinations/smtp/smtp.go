package destinations

import (
	"context"
	"log/slog"

	ghrelnoty "it.davquar/gitrelnot/internal/ghrelnoty"
)

const DestinationType string = "smtp"

type SMTPDestination struct {
	ghrelnoty.Destination
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Server   string `yaml:"server"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// TODO Implement
func (d SMTPDestination) Send(ctx context.Context, text string) error {
	slog.DebugContext(ctx, "fake email send", slog.String("from", d.From), slog.String("to", d.To), slog.String("text", text))
	return nil
}
