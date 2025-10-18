package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"fitonex/backend/internal/httpx"
	"fitonex/backend/internal/models"
	"fitonex/backend/internal/moderation"
	"fitonex/backend/internal/pagination"

	"github.com/go-chi/chi/v5"
)

// GetNearbyGyms returns gyms ordered by distance with optional caching.
func (h *Handlers) GetNearbyGyms(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	latStr := query.Get("lat")
	lngStr := query.Get("lng")
	radiusStr := query.Get("radius_km")
	limitStr := query.Get("limit")
	cursorStr := query.Get("cursor")

	if latStr == "" || lngStr == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "lat and lng parameters are required")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil || math.IsNaN(lat) || math.IsInf(lat, 0) || lat < -90 || lat > 90 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "lat must be a valid coordinate between -90 and 90")
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil || math.IsNaN(lng) || math.IsInf(lng, 0) || lng < -180 || lng > 180 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "lng must be a valid coordinate between -180 and 180")
		return
	}

	radius := 5.0
	if radiusStr != "" {
		radius, err = strconv.ParseFloat(radiusStr, 64)
		if err != nil || math.IsNaN(radius) || math.IsInf(radius, 0) || radius <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "radius_km must be a positive number")
			return
		}
		if radius > 50 {
			radius = 50
		}
	}

	limit := 20
	if limitStr != "" {
		value, err := strconv.Atoi(limitStr)
		if err != nil || value <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "limit must be a positive integer")
			return
		}
		if value > 50 {
			value = 50
		}
		limit = value
	}

	var cursor *pagination.DistanceAscCursor
	if cursorStr != "" {
		decoded, err := pagination.DecodeCursor[pagination.DistanceAscCursor](cursorStr)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid cursor")
			return
		}
		cursor = &decoded
	}

	service := h.getGymsService()
	if service == nil {
		httpx.WriteAPIError(w, httpx.NewError(http.StatusInternalServerError, httpx.ErrorCodeInternal, "gym service is not available"))
		return
	}

	cacheKey := fmt.Sprintf("NEARBY:%.4f:%.4f:%.2f:%d:%s", lat, lng, radius, limit, cursorStr)
	var page pagination.Paginated[models.NearbyGym]
	if h.cache != nil {
		if ok, _ := h.cache.GetJSON(r.Context(), cacheKey, &page); ok {
			httpx.WriteJSONWithCache(w, http.StatusOK, page, h.config.CacheTTLNearby)
			return
		}
	}

	page, err := service.GetNearby(lat, lng, radius, limit, cursor)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidLimit) {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "limit must be greater than zero")
			return
		}
		var apiErr *httpx.APIError
		if errors.As(err, &apiErr) {
			httpx.WriteAPIError(w, apiErr)
			return
		}
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to fetch nearby gyms"))
		return
	}

	if h.cache != nil {
		_ = h.cache.SetJSON(r.Context(), cacheKey, page, h.config.CacheTTLNearby)
	}

	if h.analytics != nil {
		uid, _ := userIDFromContext(r)
		h.analytics.EmitEvent(r.Context(), uid, "map_opened", map[string]any{
			"lat":    lat,
			"lng":    lng,
			"radius": radius,
		})
	}

	httpx.WriteJSONWithCache(w, http.StatusOK, page, h.config.CacheTTLNearby)
}

// GetGym returns a gym with cache awareness.
func (h *Handlers) GetGym(w http.ResponseWriter, r *http.Request) {
	gymID := chi.URLParam(r, "id")
	cacheKey := fmt.Sprintf("GYM:%s", gymID)
	var gym models.Gym
	if h.cache != nil {
		if ok, _ := h.cache.GetJSON(r.Context(), cacheKey, &gym); ok {
			httpx.WriteJSONWithCache(w, http.StatusOK, gym, h.config.CacheTTLGym)
			return
		}
	}

	result, err := h.store.Gyms.GetByID(gymID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, httpx.ErrorCodeNotFound, "gym not found")
		return
	}
	gym = *result

	if h.cache != nil {
		_ = h.cache.SetJSON(r.Context(), cacheKey, gym, h.config.CacheTTLGym)
	}

	httpx.WriteJSONWithCache(w, http.StatusOK, gym, h.config.CacheTTLGym)
}

// GetGymMachines returns machines for a gym.
func (h *Handlers) GetGymMachines(w http.ResponseWriter, r *http.Request) {
	gymID := chi.URLParam(r, "id")

	machines, err := h.store.Gyms.GetMachines(gymID)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to fetch gym machines"))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, machines)
}

// GetGymPrices returns price plans for a gym.
func (h *Handlers) GetGymPrices(w http.ResponseWriter, r *http.Request) {
	gymID := chi.URLParam(r, "id")

	prices, err := h.store.Gyms.GetPrices(gymID)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to fetch gym prices"))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, prices)
}

type CreateGymReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

// CreateGymReview handles moderated review creation.
func (h *Handlers) CreateGymReview(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	gymID := chi.URLParam(r, "id")

	var req CreateGymReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid request body")
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "rating must be between 1 and 5")
		return
	}

	if h.moderationEnabled {
		if err := moderation.ValidateReview(req.Comment); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
			return
		}
	}

	review, err := h.store.Gyms.CreateReview(gymID, userID, req.Rating, req.Comment)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to create review"))
		return
	}

	if h.cache != nil {
		_ = h.cache.Delete(r.Context(), fmt.Sprintf("GYM:%s", gymID))
		_ = h.cache.InvalidatePrefix(r.Context(), "NEARBY:", 200)
	}

	if h.analytics != nil {
		h.analytics.EmitEvent(r.Context(), userID, "review_created", map[string]any{
			"gym_id": gymID,
			"rating": req.Rating,
		})
	}

	httpx.WriteJSON(w, http.StatusCreated, review)
}

// GetGymReviews returns paginated reviews.
func (h *Handlers) GetGymReviews(w http.ResponseWriter, r *http.Request) {
	gymID := chi.URLParam(r, "id")
	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}
	cursor := r.URL.Query().Get("cursor")

	reviews, nextCursor, err := h.store.Gyms.GetReviews(gymID, cursor, limit)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to fetch reviews"))
		return
	}

	response := map[string]any{
		"reviews": reviews,
		"next":    nextCursor,
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}
