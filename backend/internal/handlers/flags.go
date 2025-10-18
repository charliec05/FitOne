package handlers

import (
	"net/http"

	"fitonex/backend/internal/httpx"
)

func (h *Handlers) GetFeatureFlags(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r)
	flags := map[string]bool{}
	if h.flags != nil {
		for name := range h.config.FeatureFlags {
			flags[name] = h.flags.IsEnabled(name, uid)
		}
	}
	flagsResponse := map[string]any{
		"flags": flags,
	}
	httpx.WriteJSON(w, http.StatusOK, flagsResponse)
}
