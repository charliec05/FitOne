package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"fitonex/backend/internal/httpx"
)

type createReportRequest struct {
	ObjectType string `json:"object_type"`
	ObjectID   string `json:"object_id"`
	Reason     string `json:"reason"`
}

func (h *Handlers) CreateReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
		return
	}

	if h.reportLimiter != nil {
		if decision, err := h.reportLimiter.Allow(r.Context(), userID); err == nil {
			if !decision.Allowed {
				httpx.WriteError(w, http.StatusTooManyRequests, httpx.ErrorCodeTooManyRequests, "report rate limit exceeded")
				return
			}
		}
	}

	var req createReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid request body")
		return
	}

	req.ObjectType = strings.TrimSpace(strings.ToLower(req.ObjectType))
	req.ObjectID = strings.TrimSpace(req.ObjectID)
	req.Reason = strings.TrimSpace(req.Reason)

	if req.ObjectType != "review" && req.ObjectType != "video" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "object_type must be review or video")
		return
	}
	if req.ObjectID == "" || req.Reason == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "object_id and reason are required")
		return
	}

	report, err := h.store.Moderation.Create(userID, req.ObjectType, req.ObjectID, req.Reason)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to create report"))
		return
	}

	if h.analytics != nil {
		h.analytics.EmitEvent(r.Context(), userID, "report_created", map[string]any{
			"object_type": req.ObjectType,
			"object_id":   req.ObjectID,
		})
	}

	httpx.WriteJSON(w, http.StatusCreated, report)
}
