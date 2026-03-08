package validator

import (
	"context"
	"regexp"
	"strings"
)

type SQLValidatorImpl struct {
	dangerousPatterns []*regexp.Regexp
}

func NewSQLValidator() *SQLValidatorImpl {
	patterns := []string{
		`(?i)^\s*DROP\s+`,
		`(?i)^\s*TRUNCATE\s+`,
		`(?i)^\s*DELETE\s+FROM\s+\w+\s*;?\s*$`,
		`(?i)^\s*UPDATE\s+\w+\s+SET\s+.*\s*;?\s*$`,
	}

	compiledPatterns := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err == nil {
			compiledPatterns = append(compiledPatterns, re)
		}
	}

	return &SQLValidatorImpl{
		dangerousPatterns: compiledPatterns,
	}
}

func (v *SQLValidatorImpl) Validate(ctx context.Context, instance, query string) (*ValidationResult, error) {
	query = strings.TrimSpace(query)

	for _, pattern := range v.dangerousPatterns {
		if pattern.MatchString(query) {
			return &ValidationResult{
				Allowed:   false,
				Reason:    "SQL query contains dangerous operation",
				RiskLevel: "HIGH",
			}, nil
		}
	}

	if v.isDeleteWithoutWhere(query) {
		return &ValidationResult{
			Allowed:   false,
			Reason:    "DELETE operation without WHERE clause is not allowed",
			RiskLevel: "HIGH",
		}, nil
	}

	if v.isUpdateWithoutWhere(query) {
		return &ValidationResult{
			Allowed:   false,
			Reason:    "UPDATE operation without WHERE clause is not allowed",
			RiskLevel: "HIGH",
		}, nil
	}

	return &ValidationResult{
		Allowed:   true,
		Reason:    "Query validated successfully",
		RiskLevel: "LOW",
	}, nil
}

func (v *SQLValidatorImpl) isDeleteWithoutWhere(query string) bool {
	query = strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToUpper(query), "DELETE") {
		return false
	}

	return !strings.Contains(strings.ToUpper(query), "WHERE")
}

func (v *SQLValidatorImpl) isUpdateWithoutWhere(query string) bool {
	query = strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToUpper(query), "UPDATE") {
		return false
	}

	return !strings.Contains(strings.ToUpper(query), "WHERE")
}

func (v *SQLValidatorImpl) GetQueryType(query string) string {
	query = strings.TrimSpace(query)
	upperQuery := strings.ToUpper(query)

	if strings.HasPrefix(upperQuery, "SELECT") || strings.HasPrefix(upperQuery, "SHOW") {
		return "SELECT"
	} else if strings.HasPrefix(upperQuery, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(upperQuery, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(upperQuery, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(upperQuery, "CREATE") {
		return "CREATE"
	} else if strings.HasPrefix(upperQuery, "ALTER") {
		return "ALTER"
	} else if strings.HasPrefix(upperQuery, "DROP") {
		return "DROP"
	} else if strings.HasPrefix(upperQuery, "TRUNCATE") {
		return "TRUNCATE"
	}

	return "UNKNOWN"
}

type WhitelistManagerImpl struct {
	patterns []*regexp.Regexp
	rawList  []string
}

func NewWhitelistManager() *WhitelistManagerImpl {
	return &WhitelistManagerImpl{
		patterns: make([]*regexp.Regexp, 0),
		rawList:  make([]string, 0),
	}
}

func (w *WhitelistManagerImpl) Add(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	w.patterns = append(w.patterns, re)
	w.rawList = append(w.rawList, pattern)
	return nil
}

func (w *WhitelistManagerImpl) Remove(pattern string) error {
	for i, p := range w.rawList {
		if p == pattern {
			w.rawList = append(w.rawList[:i], w.rawList[i+1:]...)
			w.patterns = append(w.patterns[:i], w.patterns[i+1:]...)
			return nil
		}
	}
	return nil
}

func (w *WhitelistManagerImpl) Match(query string) bool {
	for _, pattern := range w.patterns {
		if pattern.MatchString(query) {
			return true
		}
	}
	return false
}

func (w *WhitelistManagerImpl) GetAll() []string {
	return w.rawList
}

type BlacklistManagerImpl struct {
	patterns []*regexp.Regexp
	rawList  []string
}

func NewBlacklistManager() *BlacklistManagerImpl {
	return &BlacklistManagerImpl{
		patterns: make([]*regexp.Regexp, 0),
		rawList:  make([]string, 0),
	}
}

func (b *BlacklistManagerImpl) Add(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	b.patterns = append(b.patterns, re)
	b.rawList = append(b.rawList, pattern)
	return nil
}

func (b *BlacklistManagerImpl) Remove(pattern string) error {
	for i, p := range b.rawList {
		if p == pattern {
			b.rawList = append(b.rawList[:i], b.rawList[i+1:]...)
			b.patterns = append(b.patterns[:i], b.patterns[i+1:]...)
			return nil
		}
	}
	return nil
}

func (b *BlacklistManagerImpl) Match(query string) bool {
	for _, pattern := range b.patterns {
		if pattern.MatchString(query) {
			return true
		}
	}
	return false
}

func (b *BlacklistManagerImpl) GetAll() []string {
	return b.rawList
}
