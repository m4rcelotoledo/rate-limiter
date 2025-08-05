package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RateLimitIPRequestsPerSecond       int
	RateLimitIPBlockDurationSeconds    int
	RateLimitTokenRequestsPerSecond    int
	RateLimitTokenBlockDurationSeconds int
	RedisHost                          string
	RedisPort                          string
	RedisPassword                      string
	RedisDB                            int
	ServerPort                         string
}

func Load() (*Config, error) {
	// Carrega o arquivo .env se existir
	godotenv.Load()

	config := &Config{
		RateLimitIPRequestsPerSecond:       getEnvAsInt("RATE_LIMIT_IP_REQUESTS_PER_SECOND", 10),
		RateLimitIPBlockDurationSeconds:    getEnvAsInt("RATE_LIMIT_IP_BLOCK_DURATION_SECONDS", 300),
		RateLimitTokenRequestsPerSecond:    getEnvAsInt("RATE_LIMIT_TOKEN_REQUESTS_PER_SECOND", 100),
		RateLimitTokenBlockDurationSeconds: getEnvAsInt("RATE_LIMIT_TOKEN_BLOCK_DURATION_SECONDS", 600),
		RedisHost:                          getEnv("REDIS_HOST", "localhost"),
		RedisPort:                          getEnv("REDIS_PORT", "6379"),
		RedisPassword:                      getEnv("REDIS_PASSWORD", ""),
		RedisDB:                            getEnvAsInt("REDIS_DB", 0),
		ServerPort:                         getEnv("SERVER_PORT", "8080"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
