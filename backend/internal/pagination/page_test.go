package pagination

import (
	"errors"
	"math"
	"testing"
	"time"
)

type timeItem struct {
	ID        string
	CreatedAt time.Time
}

type distanceItem struct {
	ID        string
	DistanceM float64
}

func TestTimeDescPageHasMore(t *testing.T) {
	items := []timeItem{
		{ID: "1", CreatedAt: time.Date(2024, 5, 3, 10, 0, 0, 0, time.UTC)},
		{ID: "2", CreatedAt: time.Date(2024, 5, 2, 10, 0, 0, 0, time.UTC)},
		{ID: "3", CreatedAt: time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC)},
	}

	page, err := TimeDescPage(items, 2, func(item timeItem) TimeDescCursor {
		return TimeDescCursor{ID: item.ID, CreatedAt: item.CreatedAt}
	})
	if err != nil {
		t.Fatalf("TimeDescPage error: %v", err)
	}

	if len(page.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(page.Items))
	}
	if !page.HasMore {
		t.Fatal("expected HasMore true")
	}
	if page.NextCursor == "" {
		t.Fatal("expected NextCursor")
	}

	cursor, err := DecodeCursor[TimeDescCursor](page.NextCursor)
	if err != nil {
		t.Fatalf("DecodeCursor error: %v", err)
	}
	if cursor.ID != "2" {
		t.Fatalf("expected cursor id 2, got %s", cursor.ID)
	}
	if !cursor.CreatedAt.Equal(items[1].CreatedAt) {
		t.Fatalf("expected cursor time %s, got %s", items[1].CreatedAt, cursor.CreatedAt)
	}
}

func TestTimeDescPageNoMore(t *testing.T) {
	items := []timeItem{
		{ID: "1", CreatedAt: time.Now()},
		{ID: "2", CreatedAt: time.Now().Add(-time.Hour)},
	}

	page, err := TimeDescPage(items, 5, func(item timeItem) TimeDescCursor {
		return TimeDescCursor{ID: item.ID, CreatedAt: item.CreatedAt}
	})
	if err != nil {
		t.Fatalf("TimeDescPage error: %v", err)
	}
	if page.HasMore {
		t.Fatal("expected HasMore false")
	}
	if page.NextCursor != "" {
		t.Fatal("expected empty NextCursor")
	}
	if len(page.Items) != len(items) {
		t.Fatalf("expected %d items, got %d", len(items), len(page.Items))
	}
}

func TestTimeDescPageInvalidLimit(t *testing.T) {
	_, err := TimeDescPage([]timeItem{}, 0, func(item timeItem) TimeDescCursor {
		return TimeDescCursor{}
	})
	if !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
}

func TestTimeDescPageInvalidCursor(t *testing.T) {
	_, err := TimeDescPage([]timeItem{{ID: "", CreatedAt: time.Now()}}, 1, func(item timeItem) TimeDescCursor {
		return TimeDescCursor{ID: item.ID}
	})
	if !errors.Is(err, ErrInvalidCursorValue) {
		t.Fatalf("expected ErrInvalidCursorValue, got %v", err)
	}
}

func TestDistanceAscPageHasMore(t *testing.T) {
	items := []distanceItem{
		{ID: "a", DistanceM: 10},
		{ID: "b", DistanceM: 20},
		{ID: "c", DistanceM: 30},
	}

	page, err := DistanceAscPage(items, 2, func(item distanceItem) DistanceAscCursor {
		return DistanceAscCursor{ID: item.ID, DistanceM: item.DistanceM}
	})
	if err != nil {
		t.Fatalf("DistanceAscPage error: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(page.Items))
	}
	if !page.HasMore {
		t.Fatal("expected HasMore true")
	}

	cursor, err := DecodeCursor[DistanceAscCursor](page.NextCursor)
	if err != nil {
		t.Fatalf("DecodeCursor error: %v", err)
	}
	if cursor.ID != "b" {
		t.Fatalf("expected cursor id b, got %s", cursor.ID)
	}
	if cursor.DistanceM != 20 {
		t.Fatalf("expected cursor distance 20, got %f", cursor.DistanceM)
	}
}

func TestDistanceAscPageNoMore(t *testing.T) {
	items := []distanceItem{
		{ID: "a", DistanceM: 10},
	}

	page, err := DistanceAscPage(items, 5, func(item distanceItem) DistanceAscCursor {
		return DistanceAscCursor{ID: item.ID, DistanceM: item.DistanceM}
	})
	if err != nil {
		t.Fatalf("DistanceAscPage error: %v", err)
	}
	if page.HasMore {
		t.Fatal("expected HasMore false")
	}
	if page.NextCursor != "" {
		t.Fatal("expected empty NextCursor")
	}
}

func TestDistanceAscPageInvalidCursor(t *testing.T) {
	_, err := DistanceAscPage([]distanceItem{{ID: "a", DistanceM: math.NaN()}}, 1, func(item distanceItem) DistanceAscCursor {
		return DistanceAscCursor{ID: item.ID, DistanceM: item.DistanceM}
	})
	if !errors.Is(err, ErrInvalidCursorValue) {
		t.Fatalf("expected ErrInvalidCursorValue, got %v", err)
	}
}
