package types

// OIDC provider configuration types used by the auth service to bootstrap
// the OpenID Connect integration.
type OIDCProviderConfig struct {
	Name         string   `json:"name" yaml:"name"`
	Issuer       string   `json:"issuer" yaml:"issuer"`
	ClientID     string   `json:"client_id" yaml:"client_id"`
	ClientSecret string   `json:"client_secret" yaml:"client_secret"`
	RedirectURL  string   `json:"redirect_url" yaml:"redirect_url"`
	Scopes       []string `json:"scopes" yaml:"scopes"`
	RoleClaim    string   `json:"role_claim" yaml:"role_claim"`
}

type OIDCConfig struct {
	Providers []OIDCProviderConfig `json:"providers" yaml:"providers"`
}
