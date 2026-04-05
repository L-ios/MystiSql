package masking

import (
	"strings"
	"sync"
)

// MaskType defines the type of masking to apply to a column value.
type MaskType string

const (
	MaskTypePhone    MaskType = "phone"
	MaskTypeEmail    MaskType = "email"
	MaskTypeIDCard   MaskType = "idcard"
	MaskTypeBankCard MaskType = "bankcard"
	MaskTypeCustom   MaskType = "custom"
)

// ColumnRule defines a masking rule for a specific column pattern.
// Pattern format: "instance.table.column" where each segment can be "*" for wildcard.
// Examples: "*.users.phone", "*.orders.bank_card", "phone"
// Matching uses the last segment against column names.
type ColumnRule struct {
	Pattern  string   `json:"pattern" yaml:"pattern"`
	MaskType MaskType `json:"maskType" yaml:"maskType"`
	MaskFunc string   `json:"maskFunc,omitempty" yaml:"maskFunc,omitempty"`
}

// RoleRules defines masking rules for a specific role.
type RoleRules struct {
	Rules []ColumnRule `json:"rules" yaml:"rules"`
}

// PolicyStore stores masking rules per role with thread-safe access.
type PolicyStore struct {
	mu       sync.RWMutex
	policies map[string]*RoleRules
}

// NewPolicyStore creates a new empty PolicyStore.
func NewPolicyStore() *PolicyStore {
	return &PolicyStore{
		policies: make(map[string]*RoleRules),
	}
}

// Get returns the RoleRules for the given role, or nil if none exists.
func (ps *PolicyStore) Get(role string) *RoleRules {
	if ps == nil {
		return nil
	}
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.policies[role]
}

// Set replaces all rules for the given role.
func (ps *PolicyStore) Set(role string, rules *RoleRules) {
	if ps == nil {
		return
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.policies == nil {
		ps.policies = make(map[string]*RoleRules)
	}
	ps.policies[role] = rules
}

// AddRule adds a ColumnRule to the specified role's rules.
func (ps *PolicyStore) AddRule(role string, rule ColumnRule) {
	if ps == nil {
		return
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.policies == nil {
		ps.policies = make(map[string]*RoleRules)
	}
	if ps.policies[role] == nil {
		ps.policies[role] = &RoleRules{}
	}
	ps.policies[role].Rules = append(ps.policies[role].Rules, rule)
}

// RemoveRule removes the first rule with the matching pattern from the role.
func (ps *PolicyStore) RemoveRule(role string, pattern string) {
	if ps == nil {
		return
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()
	rr, ok := ps.policies[role]
	if !ok {
		return
	}
	for i, r := range rr.Rules {
		if r.Pattern == pattern {
			rr.Rules = append(rr.Rules[:i], rr.Rules[i+1:]...)
			return
		}
	}
}

// matchColumn checks whether a column name matches the rule's pattern.
// Pattern matching uses the last non-wildcard segment of the pattern
// against the column name. For example:
//   - "phone" matches column "phone"
//   - "*.users.phone" matches column "phone"
//   - "*.orders.bank_card" matches column "bank_card"
func (r *ColumnRule) matchColumn(columnName string) bool {
	if r.Pattern == "" {
		return false
	}
	segments := strings.Split(r.Pattern, ".")
	last := segments[len(segments)-1]
	if last == "*" {
		return true
	}
	return strings.EqualFold(last, columnName)
}
