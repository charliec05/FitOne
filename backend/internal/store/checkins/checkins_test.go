package checkins

import (
	"testing"
	"time"
)

func TestBuildCheckinStatsEmpty(t *testing.T) {
    stats := buildCheckinStats(nil)
    if stats.CurrentStreakDays != 0 || stats.LongestStreakDays != 0 || stats.LastCheckinDay != nil {
        t.Fatalf("expected zero stats, got %+v", stats)
    }
}

func TestBuildCheckinStatsContiguous(t *testing.T) {
    days := []time.Time{
        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
        time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
    }

    stats := buildCheckinStats(days)

    if stats.CurrentStreakDays != 3 {
        t.Fatalf("expected current streak 3, got %d", stats.CurrentStreakDays)
    }
    if stats.LongestStreakDays != 3 {
        t.Fatalf("expected longest streak 3, got %d", stats.LongestStreakDays)
    }
    if stats.LastCheckinDay == nil || !stats.LastCheckinDay.Equal(days[len(days)-1]) {
        t.Fatalf("unexpected last day %+v", stats.LastCheckinDay)
    }
}

func TestBuildCheckinStatsSparse(t *testing.T) {
    days := []time.Time{
        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
        time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC),
        time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
        time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
    }

    stats := buildCheckinStats(days)

    if stats.CurrentStreakDays != 1 {
        t.Fatalf("expected current streak 1, got %d", stats.CurrentStreakDays)
    }
    if stats.LongestStreakDays != 2 {
        t.Fatalf("expected longest streak 2, got %d", stats.LongestStreakDays)
    }
}

func TestBuildCheckinStatsTrailingConsecutive(t *testing.T) {
    days := []time.Time{
        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
        time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
        time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC),
        time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
    }

    stats := buildCheckinStats(days)

    if stats.CurrentStreakDays != 3 {
        t.Fatalf("expected current streak 3, got %d", stats.CurrentStreakDays)
    }
    if stats.LongestStreakDays != 3 {
        t.Fatalf("expected longest streak 3, got %d", stats.LongestStreakDays)
    }
}
