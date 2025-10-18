package analytics

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"
)

type Emitter struct {
	sink   string
	logger *slog.Logger
}

func NewEmitter(sink string, logger *slog.Logger) *Emitter {
	return &Emitter{
		sink:   sink,
		logger: logger,
	}
}

func (e *Emitter) EmitEvent(ctx context.Context, userID, name string, props map[string]any) {
	if e == nil {
		return
	}
	record := map[string]any{
		"user_id":   userID,
		"name":      name,
		"props":     props,
		"timestamp": time.Now().UTC(),
	}
	if e.sink == "stdout" {
		e.logger.InfoContext(ctx, "analytics_event", slog.Any("event", record))
		return
	}

	payload, err := json.Marshal(record)
	if err != nil {
		e.logger.ErrorContext(ctx, "analytics_encode_error", slog.String("error", err.Error()))
		return
	}
	e.logger.InfoContext(ctx, "analytics_event_raw", slog.String("payload", string(payload)))
}
