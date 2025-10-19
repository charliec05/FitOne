package handlers

import (
	"net/http"
	"strings"

	"fitonex/backend/internal/auth"
	"fitonex/backend/internal/models"
)

func (h *Handlers) optionalUser(r *http.Request) (*models.User, error) {
	if user, ok := currentUserFromContext(r); ok {
		return user, nil
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, nil
	}
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token := strings.TrimSpace(authHeader[7:])
		userID, err := auth.ValidateToken(token, h.config.JWTSecret)
		if err != nil {
			return nil, err
		}
		if h.store == nil || h.store.Users == nil {
			return nil, nil
		}
		return h.store.Users.GetByID(userID)
	}
	return nil, nil
}

func (h *Handlers) userIsPremium(userID string) bool {
	if userID == "" {
		return false
	}
	if h.store == nil || h.store.Users == nil {
		return false
	}
	premium, err := h.store.Users.IsPremium(userID)
	return err == nil && premium
}
