package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/thomas-illiet/KubeCoder/backend/internal/config"
)

var ErrUnauthenticated = errors.New("authentication required")

type Identity struct {
	Issuer      string
	Subject     string
	Username    string
	DisplayName string
	Email       string
}

type Verifier interface {
	Verify(context.Context, string) (Identity, error)
}

type OIDCVerifier struct {
	verifier *oidc.IDTokenVerifier
	timeout  time.Duration
}

type claims struct {
	Subject           string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	Email             string `json:"email"`
}

// New discovers the OIDC provider and creates an access-token verifier.
func New(ctx context.Context, cfg config.OIDCConfig) (*OIDCVerifier, error) {
	discoveryCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()
	discoveryURL := cfg.DiscoveryURL
	if discoveryURL == "" {
		discoveryURL = cfg.Issuer
	} else if discoveryURL != cfg.Issuer {
		discoveryCtx = oidc.InsecureIssuerURLContext(discoveryCtx, cfg.Issuer)
	}
	provider, err := oidc.NewProvider(discoveryCtx, discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	return &OIDCVerifier{verifier: provider.Verifier(&oidc.Config{ClientID: cfg.Audience}), timeout: cfg.Timeout}, nil
}

// Verify validates an access token and maps its claims to an Identity.
func (v *OIDCVerifier) Verify(ctx context.Context, rawToken string) (Identity, error) {
	verifyCtx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()
	token, err := v.verifier.Verify(verifyCtx, rawToken)
	if err != nil {
		return Identity{}, fmt.Errorf("verify access token: %w", ErrUnauthenticated)
	}
	var data claims
	if err := token.Claims(&data); err != nil || data.Subject == "" {
		return Identity{}, fmt.Errorf("decode access token claims: %w", ErrUnauthenticated)
	}
	displayName := data.Name
	if displayName == "" {
		displayName = data.PreferredUsername
	}
	return Identity{
		Issuer: token.Issuer, Subject: data.Subject, Username: data.PreferredUsername,
		DisplayName: displayName, Email: data.Email,
	}, nil
}

// BearerToken extracts a Bearer token from an Authorization header.
func BearerToken(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", ErrUnauthenticated
	}
	return parts[1], nil
}

// FromRequest authenticates an HTTP request with the supplied verifier.
func FromRequest(ctx context.Context, verifier Verifier, request *http.Request) (Identity, error) {
	token, err := BearerToken(request.Header.Get("Authorization"))
	if err != nil {
		return Identity{}, err
	}
	return verifier.Verify(ctx, token)
}
