package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/m4rcelotoledo/rate-limiter/internal/config"
	"github.com/m4rcelotoledo/rate-limiter/internal/limiter"
	"github.com/m4rcelotoledo/rate-limiter/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiterIntegration(t *testing.T) {
	// Configuração para testes
	cfg := &config.Config{
		RateLimitIPRequestsPerSecond:       5,
		RateLimitIPBlockDurationSeconds:    60,
		RateLimitTokenRequestsPerSecond:    10,
		RateLimitTokenBlockDurationSeconds: 120,
		RedisHost:                          "localhost",
		RedisPort:                          "6379",
		RedisPassword:                      "",
		RedisDB:                            0,
	}

	// Inicializa storage Redis
	storage, err := storage.NewRedisStorage(
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	require.NoError(t, err)
	defer storage.Close()

	// Configura o rate limiter
	limiterConfig := &limiter.Config{
		IPRequestsPerSecond:       cfg.RateLimitIPRequestsPerSecond,
		IPBlockDurationSeconds:    cfg.RateLimitIPBlockDurationSeconds,
		TokenRequestsPerSecond:    cfg.RateLimitTokenRequestsPerSecond,
		TokenBlockDurationSeconds: cfg.RateLimitTokenBlockDurationSeconds,
	}

	rateLimiter := limiter.NewRateLimiter(storage, limiterConfig)

	t.Run("Test IP Rate Limiting", func(t *testing.T) {
		ctx := context.Background()
		ip := "192.168.1.100"

		// Primeiras 5 requisições devem ser permitidas
		for i := 0; i < 5; i++ {
			result, err := rateLimiter.CheckLimit(ctx, ip, "ip")
			require.NoError(t, err)
			assert.True(t, result.Allowed, fmt.Sprintf("Requisição %d deve ser permitida", i+1))
		}

		// 6ª requisição deve ser bloqueada
		result, err := rateLimiter.CheckLimit(ctx, ip, "ip")
		require.NoError(t, err)
		assert.False(t, result.Allowed, "6ª requisição deve ser bloqueada")
		assert.Equal(t, int64(5), result.Limit)
		assert.Equal(t, int64(0), result.Remaining)
	})

	t.Run("Test Token Rate Limiting", func(t *testing.T) {
		ctx := context.Background()
		token := "test-token-123"

		// Primeiras 10 requisições devem ser permitidas
		for i := 0; i < 10; i++ {
			result, err := rateLimiter.CheckLimit(ctx, token, "token")
			require.NoError(t, err)
			assert.True(t, result.Allowed, fmt.Sprintf("Requisição %d deve ser permitida", i+1))
		}

		// 11ª requisição deve ser bloqueada
		result, err := rateLimiter.CheckLimit(ctx, token, "token")
		require.NoError(t, err)
		assert.False(t, result.Allowed, "11ª requisição deve ser bloqueada")
		assert.Equal(t, int64(10), result.Limit)
		assert.Equal(t, int64(0), result.Remaining)
	})

	t.Run("Test Different IPs", func(t *testing.T) {
		ctx := context.Background()
		ip1 := "192.168.1.101"
		ip2 := "192.168.1.102"

		// IP1: 5 requisições permitidas
		for i := 0; i < 5; i++ {
			result, err := rateLimiter.CheckLimit(ctx, ip1, "ip")
			require.NoError(t, err)
			assert.True(t, result.Allowed)
		}

		// IP2: 5 requisições permitidas (diferente IP)
		for i := 0; i < 5; i++ {
			result, err := rateLimiter.CheckLimit(ctx, ip2, "ip")
			require.NoError(t, err)
			assert.True(t, result.Allowed)
		}

		// IP1: 6ª requisição bloqueada
		result, err := rateLimiter.CheckLimit(ctx, ip1, "ip")
		require.NoError(t, err)
		assert.False(t, result.Allowed)
	})

	t.Run("Test Different Tokens", func(t *testing.T) {
		ctx := context.Background()
		token1 := "token-abc"
		token2 := "token-def"

		// Token1: 10 requisições permitidas
		for i := 0; i < 10; i++ {
			result, err := rateLimiter.CheckLimit(ctx, token1, "token")
			require.NoError(t, err)
			assert.True(t, result.Allowed)
		}

		// Token2: 10 requisições permitidas (diferente token)
		for i := 0; i < 10; i++ {
			result, err := rateLimiter.CheckLimit(ctx, token2, "token")
			require.NoError(t, err)
			assert.True(t, result.Allowed)
		}

		// Token1: 11ª requisição bloqueada
		result, err := rateLimiter.CheckLimit(ctx, token1, "token")
		require.NoError(t, err)
		assert.False(t, result.Allowed)
	})
}

func TestTokenExtraction(t *testing.T) {
	rateLimiter := &limiter.RateLimiter{}

	t.Run("Extract Token from Header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "test-token-123")

		token := rateLimiter.ExtractTokenFromHeader(req)
		assert.Equal(t, "test-token-123", token)
	})

	t.Run("No Token in Header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)

		token := rateLimiter.ExtractTokenFromHeader(req)
		assert.Equal(t, "", token)
	})

	t.Run("Token with Spaces", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", "  test-token-456  ")

		token := rateLimiter.ExtractTokenFromHeader(req)
		assert.Equal(t, "test-token-456", token)
	})
}

func TestClientIPExtraction(t *testing.T) {
	rateLimiter := &limiter.RateLimiter{}

	t.Run("X-Forwarded-For Header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.100")
		req.RemoteAddr = "127.0.0.1:8080"

		ip := rateLimiter.GetClientIP(req)
		assert.Equal(t, "192.168.1.100", ip)
	})

	t.Run("X-Real-IP Header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("X-Real-IP", "10.0.0.1")
		req.RemoteAddr = "127.0.0.1:8080"

		ip := rateLimiter.GetClientIP(req)
		assert.Equal(t, "10.0.0.1", ip)
	})

	t.Run("RemoteAddr Fallback", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.50:8080"

		ip := rateLimiter.GetClientIP(req)
		assert.Equal(t, "192.168.1.50:8080", ip)
	})
}

// Teste de performance para verificar comportamento sob carga
func TestRateLimiterPerformance(t *testing.T) {
	cfg := &config.Config{
		RateLimitIPRequestsPerSecond:       100,
		RateLimitIPBlockDurationSeconds:    60,
		RateLimitTokenRequestsPerSecond:    1000,
		RateLimitTokenBlockDurationSeconds: 120,
		RedisHost:                          "localhost",
		RedisPort:                          "6379",
		RedisPassword:                      "",
		RedisDB:                            0,
	}

	storage, err := storage.NewRedisStorage(
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	require.NoError(t, err)
	defer storage.Close()

	limiterConfig := &limiter.Config{
		IPRequestsPerSecond:       cfg.RateLimitIPRequestsPerSecond,
		IPBlockDurationSeconds:    cfg.RateLimitIPBlockDurationSeconds,
		TokenRequestsPerSecond:    cfg.RateLimitTokenRequestsPerSecond,
		TokenBlockDurationSeconds: cfg.RateLimitTokenBlockDurationSeconds,
	}

	rateLimiter := limiter.NewRateLimiter(storage, limiterConfig)

	t.Run("Concurrent IP Requests", func(t *testing.T) {
		ctx := context.Background()
		ip := "192.168.1.200"
		concurrent := 10
		results := make(chan bool, concurrent)

		// Executa requisições concorrentes
		for i := 0; i < concurrent; i++ {
			go func() {
				result, err := rateLimiter.CheckLimit(ctx, ip, "ip")
				if err != nil {
					results <- false
					return
				}
				results <- result.Allowed
			}()
		}

		// Coleta resultados
		allowed := 0
		for i := 0; i < concurrent; i++ {
			if <-results {
				allowed++
			}
		}

		// Verifica se o número de requisições permitidas está correto
		assert.LessOrEqual(t, allowed, 100, "Número de requisições permitidas deve estar dentro do limite")
	})

	t.Run("Concurrent Token Requests", func(t *testing.T) {
		ctx := context.Background()
		token := "performance-token"
		concurrent := 50
		results := make(chan bool, concurrent)

		// Executa requisições concorrentes
		for i := 0; i < concurrent; i++ {
			go func() {
				result, err := rateLimiter.CheckLimit(ctx, token, "token")
				if err != nil {
					results <- false
					return
				}
				results <- result.Allowed
			}()
		}

		// Coleta resultados
		allowed := 0
		for i := 0; i < concurrent; i++ {
			if <-results {
				allowed++
			}
		}

		// Verifica se o número de requisições permitidas está correto
		assert.LessOrEqual(t, allowed, 1000, "Número de requisições permitidas deve estar dentro do limite")
	})
}
