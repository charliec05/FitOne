package handlers

import "net/http"

func userIDFromContext(r *http.Request) (string, bool) {
    if value := r.Context().Value("userID"); value != nil {
        if id, ok := value.(string); ok && id != "" {
            return id, true
        }
    }
    return "", false
}
