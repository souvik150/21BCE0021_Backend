package middleware

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var (
	jwtSecretKey = []byte("your_secret_key")
	limiterMap = make(map[string]*rate.Limiter)
	mu         sync.Mutex
)

const rateLimit = 100

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		tokenString := strings.Split(authHeader, "Bearer ")[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["userID"]
		log.Println("Authenticated user:", userID)
		c.Set("userID", userID)

		// Proceed with rate limiting
		if !rateLimitMiddleware(c, userID.(string)) {
			return
		}

		c.Next()
	}
}

func rateLimitMiddleware(c *gin.Context, userID string) bool {
	limiter := getLimiter(userID)

	if !limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		c.Abort()
		return false
	}

	return true
}

func getLimiter(userID string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if limiter, exists := limiterMap[userID]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(rateLimit)), rateLimit)
	limiterMap[userID] = limiter
	return limiter
}
