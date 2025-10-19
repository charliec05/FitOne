package payments

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"fitonex/backend/internal/config"
	"fitonex/backend/internal/models"

	stripe "github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
)

type CheckoutResult struct {
	URL string `json:"url"`
}

type WebhookResult struct {
	UserID       string
	PremiumUntil time.Time
}

type Provider interface {
	CreateCheckoutSession(ctx context.Context, user *models.User) (*CheckoutResult, error)
	HandleWebhook(ctx context.Context, payload []byte, signature string) (*WebhookResult, error)
}

type StripeProvider struct {
	secretKey     string
	webhookSecret string
	priceID       string
	successURL    string
	cancelURL     string
}

func NewStripeProvider(cfg *config.Config) *StripeProvider {
	if cfg.StripeSecretKey == "" || cfg.StripePriceID == "" {
		return nil
	}
	return &StripeProvider{
		secretKey:     cfg.StripeSecretKey,
		webhookSecret: cfg.StripeWebhookSecret,
		priceID:       cfg.StripePriceID,
		successURL:    cfg.StripeSuccessURL,
		cancelURL:     cfg.StripeCancelURL,
	}
}

func (p *StripeProvider) CreateCheckoutSession(ctx context.Context, user *models.User) (*CheckoutResult, error) {
	if p == nil {
		return nil, errors.New("stripe provider not configured")
	}
	stripe.Key = p.secretKey
	params := &stripe.CheckoutSessionParams{
		SuccessURL:        stripe.String(p.successURL),
		CancelURL:         stripe.String(p.cancelURL),
		Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ClientReferenceID: stripe.String(user.ID),
		CustomerEmail:     stripe.String(user.Email),
	}
	params.LineItems = []*stripe.CheckoutSessionLineItemParams{
		{Price: stripe.String(p.priceID), Quantity: stripe.Int64(1)},
	}
	session, err := session.New(params)
	if err != nil {
		return nil, err
	}
	return &CheckoutResult{URL: session.URL}, nil
}

func (p *StripeProvider) HandleWebhook(ctx context.Context, payload []byte, signature string) (*WebhookResult, error) {
	if p == nil {
		return nil, errors.New("stripe provider not configured")
	}
	var event stripe.Event
	var err error
	if p.webhookSecret != "" {
		event, err = webhook.ConstructEvent(payload, signature, p.webhookSecret)
		if err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(payload, &event); err != nil {
			return nil, err
		}
	}
	if event.Type != "checkout.session.completed" {
		return nil, nil
	}
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return nil, err
	}
	if session.ClientReferenceID == "" {
		return nil, errors.New("missing client reference")
	}
	return &WebhookResult{
		UserID:       session.ClientReferenceID,
		PremiumUntil: time.Now().UTC().Add(30 * 24 * time.Hour),
	}, nil
}
