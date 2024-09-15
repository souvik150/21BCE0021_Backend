package socket

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/souvik150/file-sharing-app/internal/cache"
	"github.com/souvik150/file-sharing-app/pkg/middleware"
)

var connectedClients = make(map[string]*websocket.Conn)
var clientsMutex sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	clientsMutex.Lock()
	connectedClients[userID] = conn
	clientsMutex.Unlock()

	if err := addClientToRedis(userID); err != nil {
		log.Printf("❌ Error storing client in Redis: %v", err)
		conn.Close()
		return
	}

	log.Printf("🔗 User connected. User ID: %s", userID)

	defer func() {
		conn.Close()
		removeClientFromRedis(userID)
		clientsMutex.Lock()
		delete(connectedClients, userID)
		clientsMutex.Unlock()
		log.Printf("🔴 User disconnected. User ID: %s", userID)
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("⚠️ Error reading WebSocket message: %v", err)
			break
		}
	}
}

func addClientToRedis(userID string) error {
	ctx := context.Background()
	redisClient := cache.GetClient()
	err := redisClient.SAdd(ctx, "connected_users", userID).Err()
	if err != nil {
		return err
	}
	log.Printf("📝 User added to Redis: %s", userID)
	return nil
}

func removeClientFromRedis(userID string) {
	ctx := context.Background()
	redisClient := cache.GetClient()
	err := redisClient.SRem(ctx, "connected_users", userID).Err()
	if err != nil {
		log.Printf("❌ Error removing user from Redis: %v", err)
	}
	log.Printf("🗑️ User removed from Redis: %s", userID)
}

func NotifyUser(userID, message string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	conn, exists := connectedClients[userID]
	if exists {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("⚠️ Error sending WebSocket message to user %s: %v", userID, err)
		} else {
			log.Printf("📨 Sent WebSocket notification to user %s", userID)
		}
	} else {
		log.Printf("🔴 No active WebSocket connection for user %s", userID)
	}
}
