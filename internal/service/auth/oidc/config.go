package oidc

// OIDCProviderConfig holds the configuration for a single OIDC provider.
// This is a minimal schema that can be extended as needed.
type OIDCProviderConfig struct {
	Name         string   `json:"name" yaml:"name"`
	Issuer       string   `json:"issuer" yaml:"issuer"`
	ClientID     string   `json:"client_id" yaml:"client_id"`
	ClientSecret string   `json:"client_secret" yaml:"client_secret"`
	RedirectURL  string   `json:"redirect_url" yaml:"redirect_url"`
	Scopes       []string `json:"scopes" yaml:"scopes"`
	// RoleClaim specifies the JWT/ID Token claim that contains roles/groups
	// (e.g., "roles" or "groups"). If empty, a best-effort mapping will be used.
	RoleClaim string `json:"role_claim" yaml:"role_claim"`
}

// OIDCConfig aggregates multiple OIDC provider configurations.
type OIDCConfig struct {
	Providers []OIDCProviderConfig `json:"providers" yaml:"providers"`
}
