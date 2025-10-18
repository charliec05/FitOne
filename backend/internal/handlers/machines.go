package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fitonex/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// SearchMachinesRequest represents the machine search request
type SearchMachinesRequest struct {
	Query    string `json:"query"`
	BodyPart string `json:"body_part"`
	Limit    int    `json:"limit"`
}

// SearchMachinesResponse represents the machine search response
type SearchMachinesResponse struct {
	Machines []models.Machine `json:"machines"`
}

// SearchMachines handles searching machines
func (h *Handlers) SearchMachines(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	query := r.URL.Query().Get("query")
	bodyPart := r.URL.Query().Get("body_part")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 20 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	machines, err := h.store.Machines.Search(query, bodyPart, limit)
	if err != nil {
		http.Error(w, "Failed to search machines", http.StatusInternalServerError)
		return
	}

	response := SearchMachinesResponse{
		Machines: machines,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMachine handles getting a specific machine
func (h *Handlers) GetMachine(w http.ResponseWriter, r *http.Request) {
	machineID := chi.URLParam(r, "id")

	machine, err := h.store.Machines.GetByID(machineID)
	if err != nil {
		http.Error(w, "Machine not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(machine)
}

// GetBodyParts handles getting all body parts
func (h *Handlers) GetBodyParts(w http.ResponseWriter, r *http.Request) {
	bodyParts, err := h.store.Machines.GetBodyParts()
	if err != nil {
		http.Error(w, "Failed to fetch body parts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bodyParts)
}
