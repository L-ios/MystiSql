package oidc

import (
	"context"

	"encoding/json"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"time"
)

// IdentityTokenInfo contains a minimal subset of claims extracted from the ID Token.
type IdentityTokenInfo struct {
	Subject string                 `json:"sub"`
	Email   string                 `json:"email"`
	Name    string                 `json:"name"`
	Claims  map[string]interface{} `json:"claims"`
}

// ValidateIDToken validates the raw ID token and returns its parsed claims.
func (c *OIDCClient) ValidateIDToken(ctx context.Context, rawIDToken string) (*IdentityTokenInfo, error) {
	idToken, err := c.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}
	claims := map[string]interface{}{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}
	// Extract common fields, with fallbacks.
	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	return &IdentityTokenInfo{Subject: sub, Email: email, Name: name, Claims: claims}, nil
}

// DecodeClaims is a helper to pretty-print the claims for debugging or logging.
func (i *IdentityTokenInfo) DecodeClaims() string {
	b, _ := json.Marshal(i.Claims)
	return string(b)
}

// Ensure some extra timestamp utility (not strictly required but helps testing).
func (i *IdentityTokenInfo) ExpiryFrom(idToken *oidc.IDToken) (time.Time, error) {
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return time.Time{}, err
	}
	if claims.Exp == 0 {
		return time.Time{}, fmt.Errorf("no exp claim in ID token")
	}
	return time.Unix(claims.Exp, 0), nil
}
