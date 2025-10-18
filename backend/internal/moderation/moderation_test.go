package moderation

import (
	"strings"
	"testing"
)

func TestValidateReview_AllowsCleanContent(t *testing.T) {
	if err := ValidateReview("Great gym with friendly staff"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateReview_BlocksProfanity(t *testing.T) {
	if err := ValidateReview("This place is full of scam"); err == nil {
		t.Fatal("expected error for blocked phrase")
	}
}

func TestValidateVideoMeta_EmojiLimit(t *testing.T) {
	bad := "Nice video " + strings.Repeat("\U0001F600", 11)
	if err := ValidateVideoMeta("Leg day", bad); err == nil {
		t.Fatal("expected error for too many emoji")
	}
}
