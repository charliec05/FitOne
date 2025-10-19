package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type GoogleProfile struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}

type GoogleVerifier struct {
	clientID string
}

func NewGoogleVerifier(clientID string) *GoogleVerifier {
	return &GoogleVerifier{clientID: clientID}
}

func (g *GoogleVerifier) Verify(ctx context.Context, token string) (*GoogleProfile, error) {
	if token == "" {
		return nil, errors.New("token required")
	}
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid token format")
	}
	payload, err := decodeSegment(parts[1])
	if err != nil {
		return nil, err
	}
	var profile GoogleProfile
	if err := json.Unmarshal(payload, &profile); err != nil {
		return nil, err
	}
	if g.clientID != "" {
		var aud struct {
			Aud string `json:"aud"`
		}
		if err := json.Unmarshal(payload, &aud); err == nil {
			if aud.Aud != g.clientID {
				return nil, fmt.Errorf("audience mismatch")
			}
		}
	}
	if profile.Email == "" || profile.Subject == "" {
		return nil, errors.New("incomplete profile")
	}
	if profile.Name == "" {
		profile.Name = profile.Email
	}
	return &profile, nil
}

func decodeSegment(seg string) ([]byte, error) {
	seg = pad(seg)
	data, err := base64.URLEncoding.DecodeString(seg)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func pad(s string) string {
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return s
}
