package observability

import (
	"context"
	"log/slog"
)

type Tracker struct {
	logger *slog.Logger
}

func NewTracker(logger *slog.Logger) *Tracker {
	return &Tracker{logger: logger}
}

func (t *Tracker) Capture(ctx context.Context, err error, fields map[string]any) {
	if err == nil || t == nil {
		return
	}
	attrs := []any{"error", err.Error()}
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	t.logger.ErrorContext(ctx, "error_captured", attrs...)
}
