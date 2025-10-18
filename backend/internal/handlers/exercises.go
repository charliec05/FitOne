package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fitonex/backend/internal/httpx"
	"fitonex/backend/internal/models"
	"fitonex/backend/internal/pagination"

	"github.com/go-chi/chi/v5"
)

const (
	defaultExerciseLimit = 20
	maxExerciseLimit     = 50
)

// CreateExerciseRequest represents the create exercise request
type CreateExerciseRequest struct {
	Day       string      `json:"day"`
	GymID     *string     `json:"gym_id,omitempty"`
	MachineID *string     `json:"machine_id,omitempty"`
	Name      string      `json:"name"`
	Sets      []models.Set `json:"sets"`
}

// CreateExercise handles creating a new exercise
func (h *Handlers) CreateExercise(w http.ResponseWriter, r *http.Request) {
    userID, ok := userIDFromContext(r)
    if !ok {
        httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
        return
    }

    var req CreateExerciseRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid request body")
        return
    }

    req.Day = strings.TrimSpace(req.Day)
    req.Name = strings.TrimSpace(req.Name)

    if req.Day == "" || req.Name == "" || len(req.Sets) == 0 {
        httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "day, name, and sets are required")
        return
    }

    day, err := time.Parse("2006-01-02", req.Day)
    if err != nil {
        httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "day must use YYYY-MM-DD format")
        return
    }

    for i, set := range req.Sets {
        if set.Reps <= 0 {
            httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "reps must be greater than 0")
            return
        }
        if set.WeightKg != nil && *set.WeightKg < 0 {
            httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "weight must be non-negative")
            return
        }
        if set.RPE != nil && (*set.RPE < 1 || *set.RPE > 10) {
            httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "rpe must be between 1 and 10")
            return
        }
    }

    var gymID *string
    if req.GymID != nil {
        trimmed := strings.TrimSpace(*req.GymID)
        if trimmed != "" {
            gymID = &trimmed
        }
    }

    var machineID *string
    if req.MachineID != nil {
        trimmed := strings.TrimSpace(*req.MachineID)
        if trimmed != "" {
            machineID = &trimmed
        }
    }

    now := time.Now().UTC()
    performedAt := day.Add(now.Sub(now.Truncate(24 * time.Hour)))

    exercise, err := h.store.Exercises.Create(userID, performedAt, gymID, machineID, req.Name, req.Sets)
    if err != nil {
        httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to create exercise"))
        return
    }

    httpx.WriteJSON(w, http.StatusCreated, exercise)
}

// GetExercises handles getting exercises for a specific day
func (h *Handlers) GetExercises(w http.ResponseWriter, r *http.Request) {
    userID, ok := userIDFromContext(r)
    if !ok {
        httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
        return
    }

    dayStr := strings.TrimSpace(r.URL.Query().Get("day"))
    if dayStr == "" {
        httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "day parameter is required")
        return
    }

    day, err := time.Parse("2006-01-02", dayStr)
    if err != nil {
        httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "day must use YYYY-MM-DD format")
        return
    }

	limit := defaultExerciseLimit
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		value, err := strconv.Atoi(limitParam)
		if err != nil || value <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "limit must be a positive integer")
			return
		}
		if value > maxExerciseLimit {
			value = maxExerciseLimit
		}
		limit = value
	}

    var cursorPtr *pagination.TimeDescCursor
    if cursorStr := strings.TrimSpace(r.URL.Query().Get("cursor")); cursorStr != "" {
        cursor, err := pagination.DecodeCursor[pagination.TimeDescCursor](cursorStr)
        if err != nil {
            httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid cursor")
            return
        }
        cursorPtr = &cursor
    }

	page, err := h.store.Exercises.ListByDay(userID, day, limit, cursorPtr)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidLimit) {
			httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "limit must be greater than zero")
			return
		}
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to fetch exercises"))
		return
	}

    httpx.WriteJSON(w, http.StatusOK, page)
}

// GetExercise handles getting a specific exercise
func (h *Handlers) GetExercise(w http.ResponseWriter, r *http.Request) {
    userID, ok := userIDFromContext(r)
    if !ok {
        httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
        return
    }

    exerciseID := chi.URLParam(r, "id")

    exercise, err := h.store.Exercises.GetByID(exerciseID, userID)
    if err != nil {
        httpx.WriteError(w, http.StatusNotFound, httpx.ErrorCodeNotFound, "exercise not found")
        return
    }

    httpx.WriteJSON(w, http.StatusOK, exercise)
}
