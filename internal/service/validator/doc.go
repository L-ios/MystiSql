package validator

import (
	"context"
)

type ValidationResult struct {
	Allowed   bool
	Reason    string
	RiskLevel string
}

type SQLValidator interface {
	Validate(ctx context.Context, instance, query string) (*ValidationResult, error)
}

type WhitelistManager interface {
	Add(pattern string) error
	Remove(pattern string) error
	Match(query string) bool
	GetAll() []string
}

type BlacklistManager interface {
	Add(pattern string) error
	Remove(pattern string) error
	Match(query string) bool
	GetAll() []string
}
