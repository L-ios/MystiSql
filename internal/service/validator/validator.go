package validator

import (
	"context"
	"regexp"
)

type SQLValidatorImpl struct {
	tokenizer *sqlTokenizer
}

func NewSQLValidator() *SQLValidatorImpl {
	return &SQLValidatorImpl{
		tokenizer: newSQLTokenizer(),
	}
}

func (v *SQLValidatorImpl) Validate(ctx context.Context, instance, query string) (*ValidationResult, error) {
	statements := v.tokenizer.Tokenize(query)

	for _, stmt := range statements {
		switch stmt.Type {
		case "DROP", "TRUNCATE":
			return &ValidationResult{
				Allowed:   false,
				Reason:    "SQL query contains dangerous operation: " + stmt.Type,
				RiskLevel: "HIGH",
			}, nil
		case "DELETE":
			if !v.tokenizer.hasKeyword(stmt.Content, "WHERE") {
				return &ValidationResult{
					Allowed:   false,
					Reason:    "DELETE operation without WHERE clause is not allowed",
					RiskLevel: "HIGH",
				}, nil
			}
		case "UPDATE":
			if !v.tokenizer.hasKeyword(stmt.Content, "WHERE") {
				return &ValidationResult{
					Allowed:   false,
					Reason:    "UPDATE operation without WHERE clause is not allowed",
					RiskLevel: "HIGH",
				}, nil
			}
		}
	}

	return &ValidationResult{
		Allowed:   true,
		Reason:    "Query validated successfully",
		RiskLevel: "LOW",
	}, nil
}

func (v *SQLValidatorImpl) isDeleteWithoutWhere(query string) bool {
	statements := v.tokenizer.Tokenize(query)
	for _, stmt := range statements {
		if stmt.Type == "DELETE" {
			return !v.tokenizer.hasKeyword(stmt.Content, "WHERE")
		}
	}
	return false
}

func (v *SQLValidatorImpl) isUpdateWithoutWhere(query string) bool {
	statements := v.tokenizer.Tokenize(query)
	for _, stmt := range statements {
		if stmt.Type == "UPDATE" {
			return !v.tokenizer.hasKeyword(stmt.Content, "WHERE")
		}
	}
	return false
}

func (v *SQLValidatorImpl) GetQueryType(query string) string {
	statements := v.tokenizer.Tokenize(query)
	if len(statements) == 0 {
		return "UNKNOWN"
	}
	stmtType := statements[0].Type
	if stmtType == "SHOW" {
		return "SELECT"
	}
	return stmtType
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
