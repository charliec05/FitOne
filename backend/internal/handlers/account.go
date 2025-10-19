package handlers

import (
	"net/http"
	"time"

	"fitonex/backend/internal/httpx"
)

func (h *Handlers) ExportAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
		return
	}
	if h.store == nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrorCodeInternal, "storage unavailable")
		return
	}
	workouts, err := h.store.Workouts.ExportByUser(userID)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to export workouts"))
		return
	}
	checkins, err := h.store.Checkins.ExportByUser(userID)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to export checkins"))
		return
	}
	videos, err := h.store.Videos.ExportByUser(userID)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to export videos"))
		return
	}
	reviews, err := h.store.Gyms.ExportReviewsByUser(userID)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to export reviews"))
		return
	}
	response := map[string]any{
		"exported_at": time.Now().UTC(),
		"workouts":   workouts,
		"checkins":   checkins,
		"videos":     videos,
		"reviews":    reviews,
	}
	httpx.WriteJSON(w, http.StatusOK, response)
}

func (h *Handlers) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
		return
	}
	if err := h.store.Workouts.DeleteByUser(userID); err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to remove workouts"))
		return
	}
	_ = h.store.Checkins.DeleteByUser(userID)
	if h.store.Videos != nil {
		_ = h.store.Videos.AnonymizeByUser(userID)
		_ = h.store.Videos.DeleteLikesByUser(userID)
	}
	if h.store.Gyms != nil {
		_ = h.store.Gyms.AnonymizeReviewsByUser(userID)
	}
	if h.store.Comments != nil {
		_ = h.store.Comments.DeleteByUser(userID)
	}
	_ = h.store.Users.ClearPremium(userID)
	if err := h.store.Users.SoftDelete(userID); err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to delete account"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
