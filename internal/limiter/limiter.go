package limiter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type RateLimiter struct {
	storage StorageStrategy
	config  *Config
}

type Config struct {
	IPRequestsPerSecond       int
	IPBlockDurationSeconds    int
	TokenRequestsPerSecond    int
	TokenBlockDurationSeconds int
}

type StorageStrategy interface {
	Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)
	Get(ctx context.Context, key string) (int64, error)
	Set(ctx context.Context, key string, value int64, expiration time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, key string) error
	Close() error
}

func NewRateLimiter(storage StorageStrategy, config *Config) *RateLimiter {
	return &RateLimiter{
		storage: storage,
		config:  config,
	}
}

type LimitResult struct {
	Allowed   bool
	Limit     int64
	Remaining int64
	ResetTime time.Time
}

func (rl *RateLimiter) CheckLimit(ctx context.Context, identifier string, limitType string) (*LimitResult, error) {
	var requestsPerSecond int
	var blockDurationSeconds int

	switch limitType {
	case "ip":
		requestsPerSecond = rl.config.IPRequestsPerSecond
		blockDurationSeconds = rl.config.IPBlockDurationSeconds
	case "token":
		requestsPerSecond = rl.config.TokenRequestsPerSecond
		blockDurationSeconds = rl.config.TokenBlockDurationSeconds
	default:
		return nil, fmt.Errorf("invalid limit type: %s", limitType)
	}

	// Chave para o storage
	key := fmt.Sprintf("rate_limit:%s:%s", limitType, identifier)

	// Duração da janela de tempo (1 segundo)
	windowDuration := time.Second
	// Duração do bloqueio
	blockDuration := time.Duration(blockDurationSeconds) * time.Second

	// Verifica se está bloqueado
	blockKey := fmt.Sprintf("block:%s:%s", limitType, identifier)
	isBlocked, err := rl.storage.Exists(ctx, blockKey)
	if err != nil {
		return nil, fmt.Errorf("error checking block status: %w", err)
	}

	if isBlocked {
		return &LimitResult{
			Allowed:   false,
			Limit:     int64(requestsPerSecond),
			Remaining: 0,
			ResetTime: time.Now().Add(blockDuration),
		}, nil
	}

	// Incrementa o contador
	currentCount, err := rl.storage.Increment(ctx, key, windowDuration)
	if err != nil {
		return nil, fmt.Errorf("error incrementing counter: %w", err)
	}

	// Verifica se excedeu o limite
	if currentCount > int64(requestsPerSecond) {
		// Define o bloqueio
		err = rl.storage.Set(ctx, blockKey, 1, blockDuration)
		if err != nil {
			return nil, fmt.Errorf("error setting block: %w", err)
		}

		return &LimitResult{
			Allowed:   false,
			Limit:     int64(requestsPerSecond),
			Remaining: 0,
			ResetTime: time.Now().Add(blockDuration),
		}, nil
	}

	return &LimitResult{
		Allowed:   true,
		Limit:     int64(requestsPerSecond),
		Remaining: int64(requestsPerSecond) - currentCount,
		ResetTime: time.Now().Add(windowDuration),
	}, nil
}

func (rl *RateLimiter) ExtractTokenFromHeader(r *http.Request) string {
	apiKey := r.Header.Get("API_KEY")
	if apiKey != "" {
		return strings.TrimSpace(apiKey)
	}
	return ""
}

func (rl *RateLimiter) GetClientIP(r *http.Request) string {
	// Verifica headers de proxy
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Client-IP"); ip != "" {
		return ip
	}

	// Retorna o IP remoto
	return r.RemoteAddr
}
