package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidCursor indicates the cursor is malformed.
var ErrInvalidCursor = errors.New("pagination: invalid cursor")

// EncodeCursor serializes the provided value as JSON and base64 encodes it.
func EncodeCursor(v any) (string, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}

	return base64.StdEncoding.EncodeToString(payload), nil
}

// DecodeCursor deserializes the provided base64 encoded JSON cursor into T.
func DecodeCursor[T any](s string) (T, error) {
	var zero T

	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return zero, fmt.Errorf("decode cursor: %w", ErrInvalidCursor)
	}

	raw, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return zero, fmt.Errorf("decode cursor: %w", ErrInvalidCursor)
	}

	if len(raw) == 0 {
		return zero, fmt.Errorf("decode cursor: %w", ErrInvalidCursor)
	}

	var cursor T
	if err := json.Unmarshal(raw, &cursor); err != nil {
		return zero, fmt.Errorf("decode cursor: %w", ErrInvalidCursor)
	}

	return cursor, nil
}
