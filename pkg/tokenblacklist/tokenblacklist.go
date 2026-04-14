package tokenblacklist

import (
	"context"
	"sync"
	"time"
)

// entry armazena um token revogado com o instante de expiração original.
type entry struct {
	expiresAt time.Time
}

// Blacklist é uma blacklist de tokens JWT segura para uso concorrente.
type Blacklist struct {
	mu     sync.RWMutex
	tokens map[string]entry
}

// New cria uma Blacklist e inicia a goroutine de limpeza periódica.
// A goroutine encerra quando ctx for cancelado — necessário para shutdown limpo.
func New(ctx context.Context, cleanupInterval time.Duration) *Blacklist {
	bl := &Blacklist{
		tokens: make(map[string]entry),
	}
	go bl.cleanupLoop(ctx, cleanupInterval)
	return bl
}

// Revoke adiciona um token à blacklist até o instante expiresAt.
// Tokens já expirados são ignorados pois não representam risco.
func (bl *Blacklist) Revoke(token string, expiresAt time.Time) {
	if time.Now().After(expiresAt) {
		return
	}
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.tokens[token] = entry{expiresAt: expiresAt}
}

// IsRevoked retorna true se o token estiver na blacklist e ainda não tiver expirado.
func (bl *Blacklist) IsRevoked(token string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	e, ok := bl.tokens[token]
	if !ok {
		return false
	}
	return time.Now().Before(e.expiresAt)
}

// cleanupLoop remove tokens expirados periodicamente.
// Encerra quando ctx.Done() for fechado, permitindo shutdown limpo do servidor.
func (bl *Blacklist) cleanupLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bl.removeExpired()
		}
	}
}

// removeExpired deleta entradas expiradas sob lock exclusivo.
func (bl *Blacklist) removeExpired() {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	now := time.Now()
	for token, e := range bl.tokens {
		if now.After(e.expiresAt) {
			delete(bl.tokens, token)
		}
	}
}
