package models

import (
	"sync"
	"time"
)

type TokenInfo struct {
	Token     string
	ExpiresAt time.Time
	UserID    uint
}

type TokenStore struct {
	tokens map[string]*TokenInfo
	mu     sync.RWMutex
}

func NewTokenStore() *TokenStore {
	return &TokenStore{
		tokens: make(map[string]*TokenInfo),
	}
}

func (ts *TokenStore) Set(token string, info *TokenInfo) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tokens[token] = info
}

func (ts *TokenStore) Get(token string) (*TokenInfo, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	info, exists := ts.tokens[token]
	return info, exists
}

func (ts *TokenStore) Delete(token string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.tokens, token)
}

func (ts *TokenStore) CleanExpired() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	now := time.Now()
	for token, info := range ts.tokens {
		if now.After(info.ExpiresAt) {
			delete(ts.tokens, token)
		}
	}
}

func (ts *TokenStore) Clean() {
	ts.CleanExpired()
}
