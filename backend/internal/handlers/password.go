package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"fitonex/backend/internal/httpx"

	"github.com/google/uuid"
)

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (h *Handlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "email required")
		return
	}
	user, err := h.store.Users.GetByEmail(req.Email)
	if err != nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	token := uuid.NewString()
	expiresAt := time.Now().UTC().Add(1 * time.Hour)
	_ = h.store.Users.CreatePasswordResetToken(user.ID, token, expiresAt)
	if h.emails != nil {
		_ = h.emails.Send(r.Context(), user.Email, "FitONEX password reset", "Use this token: "+token)
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid body")
		return
	}
	if len(req.Password) < 8 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "password too short")
		return
	}
	userID, err := h.store.Users.ConsumePasswordResetToken(req.Token)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
		return
	}
	if err := h.store.Users.UpdatePassword(userID, req.Password); err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to reset password"))
		return
	}
	w.WriteHeader(http.StatusOK)
}
