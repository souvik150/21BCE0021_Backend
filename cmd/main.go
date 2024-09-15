package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/cache"
	"github.com/souvik150/file-sharing-app/internal/config"
	"github.com/souvik150/file-sharing-app/internal/cron"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/routes"
	"github.com/souvik150/file-sharing-app/internal/socket"
	appUtils "github.com/souvik150/file-sharing-app/pkg/utils"
)

func main() {
	config.LoadConfig()
	database.Connect()
	cache.Connect()

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
		socket.HandleWebSocket(c.Writer, c.Request)
	})

	go func() {
		if err := router.Run(":8080"); err != nil {
			log.Fatalf("❌ HTTP server failed: %v", err)
		}
	}()

	select {}
}