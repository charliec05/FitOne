package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"fitonex/backend/internal/auth"
	"fitonex/backend/internal/httpx"

	"github.com/google/uuid"
)

type googleOAuthRequest struct {
	Token string `json:"token"`
}

func (h *Handlers) GoogleOAuth(w http.ResponseWriter, r *http.Request) {
	if h.oauth == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrorCodeInternal, "oauth disabled")
		return
	}
	var req googleOAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Token) == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "token required")
		return
	}
	profile, err := h.oauth.Verify(r.Context(), req.Token)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, err.Error())
		return
	}
	user, err := h.store.Users.GetByOAuth("google", profile.Subject)
	if err != nil {
		existing, findErr := h.store.Users.GetByEmail(profile.Email)
		if findErr != nil {
			tempPassword := uuid.NewString()
			existing, findErr = h.store.Users.Create(profile.Email, tempPassword, profile.Name)
			if findErr != nil {
				httpx.WriteAPIError(w, httpx.WrapError(findErr, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to create user"))
				return
			}
		}
		if err := h.store.Users.SetOAuth(existing.ID, "google", profile.Subject); err != nil {
			httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to link oauth"))
			return
		}
		user = existing
	}
	if user.DeletedAt != nil {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "account inactive")
		return
	}
	user.Name = profile.Name
	if updated, err := h.store.Users.Update(user.ID, profile.Name, profile.Email); err == nil {
		user = updated
	}
	token, err := auth.GenerateToken(user.ID, h.config.JWTSecret)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to generate token"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, AuthResponse{Token: token, User: *user})
}
