package pagination

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"time"
)

// ErrInvalidLimit indicates the provided limit is zero or negative.
var ErrInvalidLimit = errors.New("pagination: limit must be greater than zero")

// ErrInvalidCursorValue indicates extracted cursor data is incomplete.
var ErrInvalidCursorValue = errors.New("pagination: cursor value is invalid")

// Paginated represents the standard paginated response envelope.
type Paginated[T any] struct {
	Items      []T   `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// TimeDescCursor models a cursor for created_at DESC pagination.
type TimeDescCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

// DistanceAscCursor models a cursor for distance ASC pagination.
type DistanceAscCursor struct {
	DistanceM float64 `json:"distance_m"`
	ID        string  `json:"id"`
}

// ScoreDescCursor models cursor based on similarity/score desc.
type ScoreDescCursor struct {
	Score float64 `json:"score"`
	ID    string  `json:"id"`
}

// TimeDescPage builds a paginated response for time DESC ordered lists.
func TimeDescPage[T any](items []T, limit int, extractor func(item T) TimeDescCursor) (Paginated[T], error) {
	return buildPage(items, limit, func(item T) (any, error) {
		cursor := extractor(item)
		if cursor.ID == "" || cursor.CreatedAt.IsZero() {
			return nil, ErrInvalidCursorValue
		}
		cursor.CreatedAt = cursor.CreatedAt.UTC()
		return cursor, nil
	})
}

// DistanceAscPage builds a paginated response for distance ASC ordered lists.
func DistanceAscPage[T any](items []T, limit int, extractor func(item T) DistanceAscCursor) (Paginated[T], error) {
	return buildPage(items, limit, func(item T) (any, error) {
		cursor := extractor(item)
		if cursor.ID == "" || math.IsNaN(cursor.DistanceM) || math.IsInf(cursor.DistanceM, 0) {
			return nil, ErrInvalidCursorValue
		}
		return cursor, nil
	})
}

// ScoreDescPage builds a paginated response for similarity ordered lists.
func ScoreDescPage[T any](items []T, limit int, extractor func(item T) ScoreDescCursor) (Paginated[T], error) {
	return buildPage(items, limit, func(item T) (any, error) {
		cursor := extractor(item)
		if cursor.ID == "" || math.IsNaN(cursor.Score) || math.IsInf(cursor.Score, 0) {
			return nil, ErrInvalidCursorValue
		}
		return cursor, nil
	})
}

func buildPage[T any](items []T, limit int, cursorFn func(item T) (any, error)) (Paginated[T], error) {
	if limit <= 0 {
		return Paginated[T]{}, ErrInvalidLimit
	}

	total := len(items)
	if total == 0 {
		return Paginated[T]{Items: make([]T, 0), HasMore: false}, nil
	}

	hasMore := total > limit

	var trimmed []T
	if hasMore {
		trimmed = slices.Clone(items[:limit])
	} else {
		trimmed = slices.Clone(items)
	}

	page := Paginated[T]{
		Items:   trimmed,
		HasMore: hasMore,
	}

	if hasMore {
		last := trimmed[len(trimmed)-1]
		cursorValue, err := cursorFn(last)
		if err != nil {
			return Paginated[T]{}, fmt.Errorf("build page cursor: %w", err)
		}

		nextCursor, err := EncodeCursor(cursorValue)
		if err != nil {
			return Paginated[T]{}, fmt.Errorf("build page cursor encode: %w", err)
		}
		page.NextCursor = nextCursor
	}

	return page, nil
}
