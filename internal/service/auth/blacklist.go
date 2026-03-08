package auth

import (
	"sync"
	"time"
)

type BlacklistItem struct {
	Token     string
	RevokedAt time.Time
	Reason    string
}

type TokenBlacklist struct {
	mu    sync.RWMutex
	items map[string]*BlacklistItem
}

func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{
		items: make(map[string]*BlacklistItem),
	}
}

func (b *TokenBlacklist) Add(token, reason string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.items[token] = &BlacklistItem{
		Token:     token,
		RevokedAt: time.Now(),
		Reason:    reason,
	}
}

func (b *TokenBlacklist) Contains(token string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	_, exists := b.items[token]
	return exists
}

func (b *TokenBlacklist) Remove(token string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.items, token)
}

func (b *TokenBlacklist) GetAll() []*BlacklistItem {
	b.mu.RLock()
	defer b.mu.RUnlock()

	items := make([]*BlacklistItem, 0, len(b.items))
	for _, item := range b.items {
		items = append(items, item)
	}
	return items
}

func (b *TokenBlacklist) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.items = make(map[string]*BlacklistItem)
}

func (b *TokenBlacklist) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.items)
}
