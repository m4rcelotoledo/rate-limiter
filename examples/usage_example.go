package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/m4rcelotoledo/rate-limiter/internal/config"
	"github.com/m4rcelotoledo/rate-limiter/internal/limiter"
	"github.com/m4rcelotoledo/rate-limiter/internal/middleware"
	"github.com/m4rcelotoledo/rate-limiter/internal/storage"

	"github.com/gin-gonic/gin"
)

// Exemplo de como usar o rate limiter em um projeto existente
func main() {
	// 1. Carrega configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Inicializa o storage (Redis)
	redisStorage, err := storage.NewRedisStorage(
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisStorage.Close()

	// 3. Configura o rate limiter
	limiterConfig := &limiter.Config{
		IPRequestsPerSecond:       cfg.RateLimitIPRequestsPerSecond,
		IPBlockDurationSeconds:    cfg.RateLimitIPBlockDurationSeconds,
		TokenRequestsPerSecond:    cfg.RateLimitTokenRequestsPerSecond,
		TokenBlockDurationSeconds: cfg.RateLimitTokenBlockDurationSeconds,
	}

	rateLimiter := limiter.NewRateLimiter(redisStorage, limiterConfig)

	// 4. Configura o servidor web (Gin)
	router := gin.Default()

	// 5. Adiciona o middleware de rate limiting
	router.Use(middleware.RateLimiterMiddleware(rateLimiter))

	// 6. Define suas rotas normalmente
	router.GET("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Lista de usuários",
			"data":    []string{"user1", "user2", "user3"},
		})
	})

	router.POST("/api/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Usuário criado com sucesso",
		})
	})

	router.GET("/api/products", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Lista de produtos",
			"data":    []string{"produto1", "produto2"},
		})
	})

	// 7. Inicia o servidor
	log.Printf("Servidor iniciado na porta %s", cfg.ServerPort)
	log.Fatal(router.Run(":" + cfg.ServerPort))
}

// Exemplo de como usar o rate limiter diretamente (sem middleware)
func directUsageExample() {
	// Configuração
	cfg, _ := config.Load()
	redisStorage, _ := storage.NewRedisStorage(
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	defer redisStorage.Close()

	limiterConfig := &limiter.Config{
		IPRequestsPerSecond:       10,
		IPBlockDurationSeconds:    300,
		TokenRequestsPerSecond:    100,
		TokenBlockDurationSeconds: 600,
	}

	rateLimiter := limiter.NewRateLimiter(redisStorage, limiterConfig)

	// Exemplo de uso direto
	ctx := context.Background()

	// Verifica limite por token
	result, err := rateLimiter.CheckLimit(ctx, "token-abc", "token")
	if err != nil {
		log.Printf("Erro ao verificar limite por token: %v", err)
		return
	}

	if !result.Allowed {
		log.Println("Acesso negado - limite excedido")
	}
}

// Exemplo de como implementar um storage customizado
type CustomStorage struct {
	data map[string]int64
}

func NewCustomStorage() *CustomStorage {
	return &CustomStorage{
		data: make(map[string]int64),
	}
}

func (c *CustomStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	c.data[key]++
	return c.data[key], nil
}

func (c *CustomStorage) Get(ctx context.Context, key string) (int64, error) {
	return c.data[key], nil
}

func (c *CustomStorage) Set(ctx context.Context, key string, value int64, expiration time.Duration) error {
	c.data[key] = value
	return nil
}

func (c *CustomStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := c.data[key]
	return exists, nil
}

func (c *CustomStorage) Delete(ctx context.Context, key string) error {
	delete(c.data, key)
	return nil
}

func (c *CustomStorage) Close() error {
	return nil
}

// Exemplo de uso com storage customizado
func customStorageExample() {
	customStorage := NewCustomStorage()

	limiterConfig := &limiter.Config{
		IPRequestsPerSecond:       5,
		IPBlockDurationSeconds:    60,
		TokenRequestsPerSecond:    10,
		TokenBlockDurationSeconds: 120,
	}

	rateLimiter := limiter.NewRateLimiter(customStorage, limiterConfig)

	// Use o rate limiter normalmente
	ctx := context.Background()
	result, err := rateLimiter.CheckLimit(ctx, "test-ip", "ip")
	if err != nil {
		log.Printf("Erro: %v", err)
		return
	}

	if result.Allowed {
		log.Println("Acesso permitido")
	} else {
		log.Println("Acesso negado - limite excedido")
	}
}
