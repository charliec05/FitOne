package pagination

import (
	"encoding/base64"
	"testing"
	"time"
)

func TestEncodeDecodeCursor(t *testing.T) {
	input := TimeDescCursor{
		CreatedAt: time.Date(2024, 5, 1, 12, 30, 0, 0, time.UTC),
		ID:        "123e4567-e89b-12d3-a456-426614174000",
	}

	encoded, err := EncodeCursor(input)
	if err != nil {
		t.Fatalf("EncodeCursor error: %v", err)
	}

	decoded, err := DecodeCursor[TimeDescCursor](encoded)
	if err != nil {
		t.Fatalf("DecodeCursor error: %v", err)
	}

	if !decoded.CreatedAt.Equal(input.CreatedAt) {
		t.Fatalf("expected created_at %s, got %s", input.CreatedAt, decoded.CreatedAt)
	}
	if decoded.ID != input.ID {
		t.Fatalf("expected id %s, got %s", input.ID, decoded.ID)
	}
}

func TestDecodeCursorInvalidBase64(t *testing.T) {
	_, err := DecodeCursor[TimeDescCursor]("not-base64")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecodeCursorInvalidJSON(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte("{invalid-json"))
	_, err := DecodeCursor[TimeDescCursor](payload)
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestDecodeCursorEmpty(t *testing.T) {
	_, err := DecodeCursor[TimeDescCursor]("  ")
	if err == nil {
		t.Fatal("expected error for empty cursor")
	}
}
