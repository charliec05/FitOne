package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"fitonex/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// CreateWorkoutRequest represents the workout creation request
type CreateWorkoutRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Duration    int    `json:"duration"` // in minutes
	Type        string `json:"type"`
}

// UpdateWorkoutRequest represents the workout update request
type UpdateWorkoutRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	Type        string `json:"type"`
}

// WorkoutsResponse represents the workouts list response
type WorkoutsResponse struct {
	Workouts []models.Workout `json:"workouts"`
	Next     string           `json:"next,omitempty"`
}

// GetWorkouts handles getting user workouts with cursor-based pagination
func (h *Handlers) GetWorkouts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	
	// Get cursor from query parameter
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 20 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	workouts, nextCursor, err := h.store.Workouts.GetByUserID(userID, cursor, limit)
	if err != nil {
		http.Error(w, "Failed to fetch workouts", http.StatusInternalServerError)
		return
	}

	response := WorkoutsResponse{
		Workouts: workouts,
		Next:     nextCursor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateWorkout handles creating a new workout
func (h *Handlers) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var req CreateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Name == "" {
		http.Error(w, "Workout name is required", http.StatusBadRequest)
		return
	}

	workout, err := h.store.Workouts.Create(userID, req.Name, req.Description, req.Duration, req.Type)
	if err != nil {
		http.Error(w, "Failed to create workout", http.StatusInternalServerError)
		return
	}

	if h.analytics != nil {
		h.analytics.EmitEvent(r.Context(), userID, "workout_created", map[string]any{
			"workout_id": workout.ID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workout)
}

// GetWorkout handles getting a specific workout
func (h *Handlers) GetWorkout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	workoutID := chi.URLParam(r, "id")

	workout, err := h.store.Workouts.GetByID(workoutID, userID)
	if err != nil {
		http.Error(w, "Workout not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workout)
}

// UpdateWorkout handles updating a workout
func (h *Handlers) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	workoutID := chi.URLParam(r, "id")

	var req UpdateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	workout, err := h.store.Workouts.Update(workoutID, userID, req.Name, req.Description, req.Duration, req.Type)
	if err != nil {
		http.Error(w, "Failed to update workout", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workout)
}

// DeleteWorkout handles deleting a workout
func (h *Handlers) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)
	workoutID := chi.URLParam(r, "id")

	if err := h.store.Workouts.Delete(workoutID, userID); err != nil {
		http.Error(w, "Failed to delete workout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
