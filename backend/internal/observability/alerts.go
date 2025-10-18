package observability

import (
	"bytes"
	"context"
	"net/http"
)

type Alerter struct {
	webhookURL string
	client     *http.Client
}

func NewAlerter(webhookURL string) *Alerter {
	if webhookURL == "" {
		return nil
	}
	return &Alerter{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

func (a *Alerter) Notify(ctx context.Context, payload []byte) error {
	if a == nil || a.webhookURL == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.webhookURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
