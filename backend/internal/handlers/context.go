package handlers

import (
	"net/http"

	"fitonex/backend/internal/models"
)

type ctxKey string

const (
	ctxUserID ctxKey = "userID"
	ctxUser   ctxKey = "user"
    ctxTokenGeneratedAt ctxKey = "tokenIssued"
)

func userIDFromContext(r *http.Request) (string, bool) {
	if value := r.Context().Value(ctxUserID); value != nil {
		if id, ok := value.(string); ok && id != "" {
			return id, true
		}
	}
	return "", false
}

func currentUserFromContext(r *http.Request) (*models.User, bool) {
	if value := r.Context().Value(ctxUser); value != nil {
		if user, ok := value.(*models.User); ok && user != nil {
			return user, true
		}
	}
	return nil, false
}
