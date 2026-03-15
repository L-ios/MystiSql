package types

type RBACConfig struct {
	Enabled bool     `json:"enabled"`
	Roles   []string `json:"roles"`
}

type MaskingConfig struct {
	Enabled bool `json:"enabled"`
}
