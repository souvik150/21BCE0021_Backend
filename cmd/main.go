package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"

	"github.com/souvik150/file-sharing-app/internal/cache"
	"github.com/souvik150/file-sharing-app/internal/config"
	"github.com/souvik150/file-sharing-app/internal/cron"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/routes"
	"github.com/souvik150/file-sharing-app/pkg/middleware"
	appUtils "github.com/souvik150/file-sharing-app/pkg/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clientsMutex sync.Mutex
var redisClient *redis.Client

func main() {
	config.LoadConfig()
	database.Connect()
	cache.Connect()

	redisClient = cache.GetClient()
	db := database.GetDB()
	database.Migrate(db)

	cron.CleanUpExpiredLinks()

	router := gin.Default()
	router.Use(cors.Default())
	router.Use(appUtils.UnauthenticatedRateLimiterMiddleware())

	routes.SetupRoutes(router)

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "👋 Welcome to File Sharing App API (Trademarkia Assignment)",
		})
	})

	router.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c.Writer, c.Request)
	})

	go func() {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("❌ HTTP server failed: %v", err)
		}
	}()

	select {}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Failed to upgrade to WebSocket: %v", err)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		log.Println("❌ No token provided in WebSocket connection")
		conn.Close()
		return
	}

	userID, err := middleware.ExtractUserIDFromToken(token)
	if err != nil {
		log.Printf("❌ Invalid token: %v", err)
		conn.Close()
		return
	}

	if err := addClientToRedis(userID); err != nil {
		log.Printf("❌ Error storing client in Redis: %v", err)
		conn.Close()
		return
	}

	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	log.Printf("🔗 User connected. User ID: %s", userID)

	defer func() {
		conn.Close()
		removeClientFromRedis(userID)
		log.Printf("🔴 User disconnected. User ID: %s", userID)
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("⚠️ Error reading WebSocket message: %v", err)
			break
		}

		log.Printf("📨 Received WebSocket message from user %s: %s", userID, string(message))

		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("⚠️ Error writing WebSocket message: %v", err)
			break
		}
	}
}

func addClientToRedis(userID string) error {
	ctx := context.Background()
	err := redisClient.SAdd(ctx, "connected_users", userID).Err()
	if err != nil {
		return err
	}
	log.Printf("📝 User added to Redis: %s", userID)
	return nil
}

func removeClientFromRedis(userID string) {
	ctx := context.Background()
	err := redisClient.SRem(ctx, "connected_users", userID).Err()
	if err != nil {
		log.Printf("❌ Error removing user from Redis: %v", err)
	}
	log.Printf("🗑️ User removed from Redis: %s", userID)
}