package auth

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BlacklistItem represents a revoked token entry.
type BlacklistItem struct {
	Token     string    `json:"token"`
	RevokedAt time.Time `json:"revoked_at"`
	Reason    string    `json:"reason"`
	ExpiresAt time.Time `json:"expires_at"`
}

// BlacklistConfig configures the token blacklist behavior.
type BlacklistConfig struct {
	TTL             time.Duration // default 24h; 0 = no expiry for basic mode
	FilePath        string        // empty = no persistence
	CleanupInterval time.Duration // default 1m; 0 = no cleanup goroutine
}

// TokenBlacklist manages revoked tokens with optional TTL-based cleanup
// and JSON file persistence.
type TokenBlacklist struct {
	mu     sync.RWMutex
	items  map[string]*BlacklistItem
	config BlacklistConfig
	logger *zap.Logger
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewTokenBlacklist creates a basic in-memory token blacklist with no TTL,
// no persistence, and no cleanup goroutine. This exists for backward
// compatibility.
func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{
		items:  make(map[string]*BlacklistItem),
		config: BlacklistConfig{},
		logger: zap.NewNop(),
		stopCh: make(chan struct{}),
	}
}

// NewTokenBlacklistWithConfig creates a TokenBlacklist with TTL-based cleanup
// and optional JSON file persistence.
func NewTokenBlacklistWithConfig(cfg BlacklistConfig, logger *zap.Logger) (*TokenBlacklist, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	// Apply defaults.
	if cfg.TTL == 0 {
		cfg.TTL = 24 * time.Hour
	}
	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = time.Minute
	}

	bl := &TokenBlacklist{
		items:  make(map[string]*BlacklistItem),
		config: cfg,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	// Load persisted items if file path is configured.
	if cfg.FilePath != "" {
		if err := bl.loadFromFile(); err != nil {
			return nil, fmt.Errorf("failed to load blacklist from file %q: %w", cfg.FilePath, err)
		}
	}

	// Start background cleanup goroutine.
	bl.wg.Add(1)
	go bl.cleanupLoop()

	return bl, nil
}

// Add records a token as revoked with the given reason.
func (b *TokenBlacklist) Add(token, reason string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	item := &BlacklistItem{
		Token:     token,
		RevokedAt: now,
		Reason:    reason,
	}
	if b.config.TTL > 0 {
		item.ExpiresAt = now.Add(b.config.TTL)
	}

	b.items[token] = item

	if b.config.FilePath != "" {
		if err := b.appendToFile(item); err != nil {
			b.logger.Error("failed to persist blacklist item",
				zap.String("file", b.config.FilePath),
				zap.Error(err),
			)
		}
	}
}

// Contains reports whether the token is in the blacklist and not expired.
func (b *TokenBlacklist) Contains(token string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	item, exists := b.items[token]
	if !exists {
		return false
	}
	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		return false
	}
	return true
}

// Remove deletes a token from the blacklist.
func (b *TokenBlacklist) Remove(token string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.items, token)
}

// GetAll returns all blacklist items (including expired ones that haven't
// been cleaned up yet).
func (b *TokenBlacklist) GetAll() []*BlacklistItem {
	b.mu.RLock()
	defer b.mu.RUnlock()

	items := make([]*BlacklistItem, 0, len(b.items))
	for _, item := range b.items {
		items = append(items, item)
	}
	return items
}

// Clear removes all items from the blacklist.
func (b *TokenBlacklist) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.items = make(map[string]*BlacklistItem)

	if b.config.FilePath != "" {
		if err := b.rewriteFile(); err != nil {
			b.logger.Error("failed to clear blacklist file",
				zap.String("file", b.config.FilePath),
				zap.Error(err),
			)
		}
	}
}

// Size returns the number of items currently in the blacklist.
func (b *TokenBlacklist) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.items)
}

// Stop gracefully shuts down the cleanup goroutine and waits for it to
// finish.
func (b *TokenBlacklist) Stop() {
	close(b.stopCh)
	b.wg.Wait()
}

// cleanupLoop periodically removes expired items from the blacklist.
func (b *TokenBlacklist) cleanupLoop() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			return
		case <-ticker.C:
			b.removeExpired()
		}
	}
}

// removeExpired removes all expired items and rewrites the persistence file
// if configured.
func (b *TokenBlacklist) removeExpired() {
	b.mu.Lock()
	now := time.Now()
	changed := false
	for token, item := range b.items {
		if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
			delete(b.items, token)
			changed = true
		}
	}
	b.mu.Unlock()

	if changed && b.config.FilePath != "" {
		if err := b.rewriteFile(); err != nil {
			b.logger.Error("failed to rewrite blacklist file after cleanup",
				zap.String("file", b.config.FilePath),
				zap.Error(err),
			)
		}
	}
}

// loadFromFile reads the blacklist from a JSON Lines file.
func (b *TokenBlacklist) loadFromFile() error {
	f, err := os.Open(b.config.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open blacklist file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	now := time.Now()
	loaded := 0
	expired := 0

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var item BlacklistItem
		if err := json.Unmarshal(line, &item); err != nil {
			b.logger.Warn("skipping malformed blacklist entry",
				zap.String("file", b.config.FilePath),
				zap.Error(err),
			)
			continue
		}
		// Skip expired items during load.
		if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
			expired++
			continue
		}
		b.items[item.Token] = &item
		loaded++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read blacklist file: %w", err)
	}

	b.logger.Info("loaded blacklist from file",
		zap.String("file", b.config.FilePath),
		zap.Int("loaded", loaded),
		zap.Int("expired_skipped", expired),
	)
	return nil
}

// appendToFile appends a single BlacklistItem as a JSON line to the file.
func (b *TokenBlacklist) appendToFile(item *BlacklistItem) error {
	f, err := os.OpenFile(b.config.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open blacklist file for append: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal blacklist item: %w", err)
	}

	if _, err := fmt.Fprintln(f, string(data)); err != nil {
		return fmt.Errorf("write blacklist item: %w", err)
	}

	return nil
}

// rewriteFile writes all current items to the file, replacing its contents.
func (b *TokenBlacklist) rewriteFile() error {
	f, err := os.Create(b.config.FilePath)
	if err != nil {
		return fmt.Errorf("create blacklist file: %w", err)
	}
	defer f.Close()

	for _, item := range b.items {
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("marshal blacklist item: %w", err)
		}
		if _, err := fmt.Fprintln(f, string(data)); err != nil {
			return fmt.Errorf("write blacklist item: %w", err)
		}
	}

	return nil
}
