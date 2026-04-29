package validator

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type ValidatorService struct {
	validator    *SQLValidatorImpl
	astValidator *ASTValidator
	whitelist    *WhitelistManagerImpl
	blacklist    *BlacklistManagerImpl
	logger       *zap.Logger
	mu           sync.RWMutex
}

func NewValidatorService(logger *zap.Logger) *ValidatorService {
	return NewValidatorServiceWithConfig(logger, false)
}

// NewValidatorServiceWithConfig creates a ValidatorService with the option to enable AST-based validation.
func NewValidatorServiceWithConfig(logger *zap.Logger, useParser bool) *ValidatorService {
	s := &ValidatorService{
		validator: NewSQLValidator(),
		whitelist: NewWhitelistManager(),
		blacklist: NewBlacklistManager(),
		logger:    logger,
	}

	if useParser {
		s.astValidator = NewASTValidator(s.validator, logger)
	}

	return s
}

func (s *ValidatorService) Validate(ctx context.Context, instance, query string) (*ValidationResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.blacklist.Match(query) {
		s.logger.Warn("Query blocked by blacklist",
			zap.String("instance", instance),
			zap.String("query", truncateQuery(query)),
		)
		return &ValidationResult{
			Allowed:   false,
			Reason:    "Query matches blacklist pattern",
			RiskLevel: "HIGH",
		}, nil
	}

	if s.whitelist.Match(query) {
		s.logger.Debug("Query allowed by whitelist",
			zap.String("instance", instance),
			zap.String("query", truncateQuery(query)),
		)
		return &ValidationResult{
			Allowed:   true,
			Reason:    "Query matches whitelist pattern",
			RiskLevel: "LOW",
		}, nil
	}

	if s.astValidator != nil {
		return s.astValidator.Validate(ctx, instance, query)
	}
	return s.validator.Validate(ctx, instance, query)
}

func (s *ValidatorService) AddToWhitelist(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.whitelist.Add(pattern)
	if err != nil {
		return err
	}

	s.logger.Info("Added pattern to whitelist", zap.String("pattern", pattern))
	return nil
}

func (s *ValidatorService) RemoveFromWhitelist(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.whitelist.Remove(pattern)
	if err != nil {
		return err
	}

	s.logger.Info("Removed pattern from whitelist", zap.String("pattern", pattern))
	return nil
}

func (s *ValidatorService) GetWhitelist() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.whitelist.GetAll()
}

func (s *ValidatorService) AddToBlacklist(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.blacklist.Add(pattern)
	if err != nil {
		return err
	}

	s.logger.Info("Added pattern to blacklist", zap.String("pattern", pattern))
	return nil
}

func (s *ValidatorService) RemoveFromBlacklist(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.blacklist.Remove(pattern)
	if err != nil {
		return err
	}

	s.logger.Info("Removed pattern from blacklist", zap.String("pattern", pattern))
	return nil
}

func (s *ValidatorService) GetBlacklist() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.blacklist.GetAll()
}

func (s *ValidatorService) UpdateWhitelist(patterns []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.whitelist = NewWhitelistManager()
	for _, pattern := range patterns {
		if err := s.whitelist.Add(pattern); err != nil {
			s.logger.Error("Failed to add pattern to whitelist",
				zap.String("pattern", pattern),
				zap.Error(err),
			)
			return err
		}
	}

	s.logger.Info("Updated whitelist", zap.Int("count", len(patterns)))
	return nil
}

func (s *ValidatorService) UpdateBlacklist(patterns []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blacklist = NewBlacklistManager()
	for _, pattern := range patterns {
		if err := s.blacklist.Add(pattern); err != nil {
			s.logger.Error("Failed to add pattern to blacklist",
				zap.String("pattern", pattern),
				zap.Error(err),
			)
			return err
		}
	}

	s.logger.Info("Updated blacklist", zap.Int("count", len(patterns)))
	return nil
}

func truncateQuery(query string) string {
	if len(query) <= 100 {
		return query
	}
	return query[:100] + "..."
}
