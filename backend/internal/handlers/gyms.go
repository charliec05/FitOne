package handlers

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"

	"fitonex/backend/internal/httpx"
	"fitonex/backend/internal/models"
	"fitonex/backend/internal/pagination"

	"github.com/go-chi/chi/v5"
)

// GetNearbyGyms handles getting nearby gyms
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
			radius = 50 // clamp to reasonable radius to avoid heavy queries
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

	httpx.WriteJSON(w, http.StatusOK, page)
}

// GetGym handles getting a specific gym
func (h *Handlers) GetGym(w http.ResponseWriter, r *http.Request) {
	gymID := chi.URLParam(r, "id")

	gym, err := h.store.Gyms.GetByID(gymID)
	if err != nil {
		http.Error(w, "Gym not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gym)
}

// GetGymMachines handles getting machines for a gym
func (h *Handlers) GetGymMachines(w http.ResponseWriter, r *http.Request) {
	gymID := chi.URLParam(r, "id")

	machines, err := h.store.Gyms.GetMachines(gymID)
	if err != nil {
		http.Error(w, "Failed to fetch gym machines", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(machines)
}

// GetGymPrices handles getting prices for a gym
func (h *Handlers) GetGymPrices(w http.ResponseWriter, r *http.Request) {
	gymID := chi.URLParam(r, "id")

	prices, err := h.store.Gyms.GetPrices(gymID)
	if err != nil {
		http.Error(w, "Failed to fetch gym prices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prices)
}

// CreateGymReviewRequest represents the create review request
type CreateGymReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

// CreateGymReview handles creating a gym review
func (h *Handlers) CreateGymReview(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	gymID := chi.URLParam(r, "id")

	var req CreateGymReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Rating < 1 || req.Rating > 5 {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	review, err := h.store.Gyms.CreateReview(gymID, userID, req.Rating, req.Comment)
	if err != nil {
		http.Error(w, "Failed to create review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}

// GetGymReviews handles getting reviews for a gym
func (h *Handlers) GetGymReviews(w http.ResponseWriter, r *http.Request) {
	gymID := chi.URLParam(r, "id")
	
	// Get pagination parameters
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 20 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	reviews, nextCursor, err := h.store.Gyms.GetReviews(gymID, cursor, limit)
	if err != nil {
		http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"reviews": reviews,
		"next":    nextCursor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
