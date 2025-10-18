package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"fitonex/backend/internal/httpx"
	"fitonex/backend/internal/pagination"
)

func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	searchType := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("type")))
	if searchType == "" {
		searchType = "gym"
	}
	mode := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("mode")))
	prefix := mode == "prefix"

	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}

	var cursor *pagination.ScoreDescCursor
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		value, err := pagination.DecodeCursor[pagination.ScoreDescCursor](raw)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid cursor")
			return
		}
		cursor = &value
	}

	if query == "" {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"items":    []any{},
			"has_more": false,
		})
		return
	}

	switch searchType {
	case "gym":
		page, err := h.store.Gyms.Search(query, limit, cursor, prefix)
		if err != nil {
			httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to search gyms"))
			return
		}
		h.emitSearchEvent(r, searchType, query, prefix)
		httpx.WriteJSON(w, http.StatusOK, page)
	case "machine":
		page, err := h.store.Machines.SearchWithScores(query, limit, cursor, prefix)
		if err != nil {
			httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to search machines"))
			return
		}
		h.emitSearchEvent(r, searchType, query, prefix)
		httpx.WriteJSON(w, http.StatusOK, page)
	default:
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "type must be gym or machine")
	}
}

func (h *Handlers) emitSearchEvent(r *http.Request, searchType, query string, prefix bool) {
	if h.analytics == nil {
		return
	}
	uid, _ := userIDFromContext(r)
	h.analytics.EmitEvent(r.Context(), uid, "search_performed", map[string]any{
		"type":   searchType,
		"query":  query,
		"prefix": prefix,
	})
}
