package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"fitonex/backend/internal/httpx"
	"fitonex/backend/internal/models"
)

// CheckinToday handles creating a check-in for today
func (h *Handlers) CheckinToday(w http.ResponseWriter, r *http.Request) {
    userID, ok := userIDFromContext(r)
    if !ok {
        httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
        return
    }

    checkin, inserted, err := h.store.Checkins.CreateToday(userID)
    if err != nil {
        httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to record check-in"))
        return
    }

    status := http.StatusOK
    if inserted {
        status = http.StatusCreated
    }

    if h.analytics != nil {
        h.analytics.EmitEvent(r.Context(), userID, "checkin_done", map[string]any{
            "inserted": inserted,
        })
    }

    httpx.WriteJSON(w, status, struct {
        Checkin  models.Checkin `json:"checkin"`
        Inserted bool           `json:"inserted"`
    }{Checkin: *checkin, Inserted: inserted})
}

// GetCheckinStats handles getting check-in statistics
func (h *Handlers) GetCheckinStats(w http.ResponseWriter, r *http.Request) {
    userID, ok := userIDFromContext(r)
    if !ok {
        httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
        return
    }

    stats, err := h.store.Checkins.GetStats(userID)
    if err != nil {
        httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to fetch check-in stats"))
        return
    }

	httpx.WriteJSON(w, http.StatusOK, stats)
}

func (h *Handlers) GetTopStreaks(w http.ResponseWriter, r *http.Request) {
	periodParam := r.URL.Query().Get("period")
	period := 7 * 24 * time.Hour
	if periodParam != "" {
		if strings.HasSuffix(periodParam, "d") {
			if days, err := strconv.Atoi(strings.TrimSuffix(periodParam, "d")); err == nil && days > 0 {
				period = time.Duration(days) * 24 * time.Hour
			}
		}
	}
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}
	entries, err := h.store.Checkins.TopStreaks(period, limit)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to fetch leaderboard"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"entries": entries,
	})
}
