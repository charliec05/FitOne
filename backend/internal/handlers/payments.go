package handlers

import (
	"io"
	"net/http"
	"time"

	"fitonex/backend/internal/httpx"
)

func (h *Handlers) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
    if h.payments == nil || h.store == nil || h.store.Users == nil {
        httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrorCodeInternal, "payments disabled")
        return
    }
	user, ok := currentUserFromContext(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.ErrorCodeUnauthorized, "unauthorized")
		return
	}
	premium, err := h.store.Users.IsPremium(user.ID)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to load premium status"))
		return
	}
	if premium {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "already premium")
		return
	}
	result, err := h.payments.CreateCheckoutSession(r.Context(), user)
	if err != nil {
		httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to create checkout session"))
		return
	}
	if h.analytics != nil {
		h.analytics.EmitEvent(r.Context(), user.ID, "checkout_started", map[string]any{"url": result.URL})
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handlers) HandlePaymentsWebhook(w http.ResponseWriter, r *http.Request) {
	if h.payments == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.ErrorCodeInternal, "payments disabled")
		return
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, "invalid payload")
		return
	}
	signature := r.Header.Get("Stripe-Signature")
	result, err := h.payments.HandleWebhook(r.Context(), payload, signature)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.ErrorCodeBadRequest, err.Error())
		return
	}
    if result != nil && h.store != nil && h.store.Users != nil {
        if result.PremiumUntil.IsZero() {
            result.PremiumUntil = time.Now().UTC().Add(30 * 24 * time.Hour)
        }
        if err := h.store.Users.SetPremium(result.UserID, result.PremiumUntil); err != nil {
            httpx.WriteAPIError(w, httpx.WrapError(err, http.StatusInternalServerError, httpx.ErrorCodeInternal, "failed to update premium"))
            return
        }
    }
	w.WriteHeader(http.StatusOK)
}
