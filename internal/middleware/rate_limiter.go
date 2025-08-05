package middleware

import (
	"net/http"
	"time"

	"github.com/m4rcelotoledo/rate-limiter/internal/limiter"

	"github.com/gin-gonic/gin"
)

func RateLimiterMiddleware(rateLimiter *limiter.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Extrai o token do header
		token := rateLimiter.ExtractTokenFromHeader(c.Request)

		// Se há token, verifica limite por token (tem prioridade sobre IP)
		if token != "" {
			result, err := rateLimiter.CheckLimit(ctx, token, "token")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
				return
			}

			if !result.Allowed {
				c.Header("X-RateLimit-Limit", string(rune(result.Limit)))
				c.Header("X-RateLimit-Remaining", string(rune(result.Remaining)))
				c.Header("X-RateLimit-Reset", result.ResetTime.Format(time.RFC3339))

				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
				})
				c.Abort()
				return
			}

			// Adiciona headers de rate limit
			c.Header("X-RateLimit-Limit", string(rune(result.Limit)))
			c.Header("X-RateLimit-Remaining", string(rune(result.Remaining)))
			c.Header("X-RateLimit-Reset", result.ResetTime.Format(time.RFC3339))

			c.Next()
			return
		}

		// Se não há token, verifica limite por IP
		clientIP := rateLimiter.GetClientIP(c.Request)
		result, err := rateLimiter.CheckLimit(ctx, clientIP, "ip")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}

		if !result.Allowed {
			c.Header("X-RateLimit-Limit", string(rune(result.Limit)))
			c.Header("X-RateLimit-Remaining", string(rune(result.Remaining)))
			c.Header("X-RateLimit-Reset", result.ResetTime.Format(time.RFC3339))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
			})
			c.Abort()
			return
		}

		// Adiciona headers de rate limit
		c.Header("X-RateLimit-Limit", string(rune(result.Limit)))
		c.Header("X-RateLimit-Remaining", string(rune(result.Remaining)))
		c.Header("X-RateLimit-Reset", result.ResetTime.Format(time.RFC3339))

		c.Next()
	}
}
