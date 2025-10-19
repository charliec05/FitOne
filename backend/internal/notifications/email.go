package notifications

import (
	"context"
	"log/slog"
)

type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type LoggerEmailSender struct {
	from   string
	logger *slog.Logger
}

func NewLoggerEmailSender(from string, logger *slog.Logger) *LoggerEmailSender {
	return &LoggerEmailSender{from: from, logger: logger}
}

func (s *LoggerEmailSender) Send(ctx context.Context, to, subject, body string) error {
	if s.logger != nil {
		s.logger.InfoContext(ctx, "email_send", slog.String("to", to), slog.String("subject", subject))
	}
	return nil
}
