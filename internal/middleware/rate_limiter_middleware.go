package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter mantém um rate limiter individual por endereço IP.
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimiterStore armazena e gerencia os limiters por IP com limpeza periódica.
type rateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	r        rate.Limit
	burst    int
}

// newRateLimiterStore cria um store e inicia a goroutine de limpeza.
// A goroutine encerra quando ctx for cancelado.
func newRateLimiterStore(ctx context.Context, r rate.Limit, burst int) *rateLimiterStore {
	s := &rateLimiterStore{
		limiters: make(map[string]*ipLimiter),
		r:        r,
		burst:    burst,
	}
	go s.cleanupLoop(ctx)
	return s
}

// get retorna o limiter do IP, criando um novo se necessário.
func (s *rateLimiterStore) get(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.limiters[ip]
	if !ok {
		entry = &ipLimiter{limiter: rate.NewLimiter(s.r, s.burst)}
		s.limiters[ip] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

// cleanupLoop remove IPs inativos há mais de 10 minutos a cada 5 minutos.
// Encerra quando ctx.Done() for fechado.
func (s *rateLimiterStore) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.removeInactive()
		}
	}
}

// removeInactive deleta entradas inativas sob lock exclusivo.
func (s *rateLimiterStore) removeInactive() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ip, entry := range s.limiters {
		if time.Since(entry.lastSeen) > 10*time.Minute {
			delete(s.limiters, ip)
		}
	}
}

// RateLimiter retorna um middleware Gin que limita requisições por IP.
// ctx deve ser o contexto raiz da aplicação para garantir shutdown limpo.
//
// r define quantas requisições por segundo são permitidas.
// burst define o pico máximo de requisições em rajada.
func RateLimiter(ctx context.Context, r rate.Limit, burst int) gin.HandlerFunc {
	store := newRateLimiterStore(ctx, r, burst)

	return func(c *gin.Context) {
		limiter := store.get(c.ClientIP())
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "muitas requisições, tente novamente em instantes",
			})
			return
		}
		c.Next()
	}
}
