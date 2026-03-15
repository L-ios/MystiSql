package oidc

import (
	"context"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCClient is a lightweight wrapper around an OIDC provider, its verifier,
// and the OAuth2 configuration used to initiate the Authorization Code flow.
type OIDCClient struct {
	Name        string
	Provider    *oidc.Provider
	Verifier    *oidc.IDTokenVerifier
	OAuthConfig *oauth2.Config
	// UserInfoURL is the endpoint used to fetch additional user information
	// after authentication. It is optional and may be empty for some providers.
	UserInfoURL string
	RoleClaim   string
}

// NewOIDCClient creates a new OIDC client for a given provider issuer.
// It fetches the provider metadata from the issuer and prepares the verifier
// and OAuth2 configuration for the Authorization Code flow.
func NewOIDCClient(ctx context.Context, issuer, clientID, clientSecret, redirectURL string, scopes []string, name string, roleClaim string) (*OIDCClient, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       append(scopes, oidc.ScopeOpenID),
	}

	// UserInfo endpoint is provider-specific. We keep it optional.
	userInfoURL := ""

	return &OIDCClient{
		Name:        name,
		Provider:    provider,
		Verifier:    verifier,
		OAuthConfig: oauthCfg,
		UserInfoURL: userInfoURL,
		RoleClaim:   roleClaim,
	}, nil
}

// Ensure a small HTTP utility to satisfy future HTTP calls in tests without
// forcing a live network fetch during initialization.
func mustHTTPClient() *http.Client {
	return &http.Client{Timeout: 5 * time.Second}
}
