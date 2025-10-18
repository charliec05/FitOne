package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fitonex/backend/internal/httpx"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// HealthCheck handles the /healthz endpoint
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type HealthDetailsResponse struct {
	Status   string         `json:"status"`
	Services map[string]any `json:"services"`
}

func (h *Handlers) HealthDetails(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp := h.healthDetailsSnapshot(ctx)
	status := http.StatusOK
	if resp.Status != "ok" {
		status = http.StatusServiceUnavailable
	}
	httpx.WriteJSON(w, status, resp)
}

func (h *Handlers) HealthDetailsCheck(ctx context.Context) error {
	resp := h.healthDetailsSnapshot(ctx)
	if resp.Status == "ok" {
		return nil
	}
	return fmt.Errorf("health degraded")
}

func (h *Handlers) healthDetailsSnapshot(ctx context.Context) HealthDetailsResponse {
	services := map[string]any{}
	status := "ok"

	if err := h.store.Ping(); err != nil {
		services["database"] = map[string]any{"status": "error", "error": err.Error()}
		status = "degraded"
	} else {
		services["database"] = map[string]any{"status": "ok"}
	}

	if h.cache != nil {
		if err := h.cache.Ping(ctx); err != nil {
			services["redis"] = map[string]any{"status": "error", "error": err.Error()}
			status = "degraded"
		} else {
			services["redis"] = map[string]any{"status": "ok"}
		}
	} else {
		services["redis"] = map[string]any{"status": "unconfigured"}
	}

	if h.storage != nil {
		if err := h.storage.Ping(ctx); err != nil {
			services["s3"] = map[string]any{"status": "error", "error": err.Error()}
			status = "degraded"
		} else {
			services["s3"] = map[string]any{"status": "ok"}
		}
	} else {
		services["s3"] = map[string]any{"status": "unconfigured"}
	}

	return HealthDetailsResponse{Status: status, Services: services}
}
