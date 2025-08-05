package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/m4rcelotoledo/rate-limiter/internal/config"
	"github.com/m4rcelotoledo/rate-limiter/internal/limiter"
	"github.com/m4rcelotoledo/rate-limiter/internal/middleware"
	"github.com/m4rcelotoledo/rate-limiter/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// Carrega configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Inicializa o storage Redis
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

	// Configura o rate limiter
	limiterConfig := &limiter.Config{
		IPRequestsPerSecond:       cfg.RateLimitIPRequestsPerSecond,
		IPBlockDurationSeconds:    cfg.RateLimitIPBlockDurationSeconds,
		TokenRequestsPerSecond:    cfg.RateLimitTokenRequestsPerSecond,
		TokenBlockDurationSeconds: cfg.RateLimitTokenBlockDurationSeconds,
	}

	rateLimiter := limiter.NewRateLimiter(redisStorage, limiterConfig)

	// Configura o servidor Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Adiciona o middleware de rate limiting
	router.Use(middleware.RateLimiterMiddleware(rateLimiter))

	// Rotas de exemplo
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Rate Limiter API",
			"status":  "running",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Request processed successfully",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Configura o servidor HTTP
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	// Canal para receber sinais de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Inicia o servidor em uma goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Aguarda sinal de interrupção
	<-quit
	log.Println("Shutting down server...")

	// Contexto com timeout para shutdown graceful
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
