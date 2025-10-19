package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"fitonex/backend/internal/httpx"
	"fitonex/backend/internal/pagination"

	"github.com/go-chi/chi/v5"
)

type createCommentRequest struct {
	Text string `json:"text"`
}

func (h *Handlers) CreateVideoComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
		return
	}
	videoID := chi.URLParam(r, "id")
	if videoID == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "missing video id")
		return
	}
	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid body")
		return
	}
	text := strings.TrimSpace(req.Text)
	if text == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "text required")
		return
	}
	if len([]rune(text)) > 500 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "comment too long")
		return
	}
	comment, err := h.store.Comments.CreateComment(videoID, userID, text)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to create comment"))
		return
	}
	if h.analytics != nil {
		h.analytics.EmitEvent(r.Context(), userID, "comment_created", map[string]any{"video_id": videoID})
	}
	httpx.WriteJSON(w, http.StatusCreated, comment)
}

func (h *Handlers) ListVideoComments(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	if videoID == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "missing video id")
		return
	}
	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	var cursor *pagination.TimeDescCursor
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		value, err := pagination.DecodeCursor[pagination.TimeDescCursor](raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid cursor")
			return
		}
		cursor = &value
	}
	page, err := h.store.Comments.ListByVideo(videoID, limit, cursor)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to list comments"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, page)
}
