package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/cache"
	"github.com/souvik150/file-sharing-app/internal/config"
	"github.com/souvik150/file-sharing-app/internal/cron"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/routes"
	appUtils "github.com/souvik150/file-sharing-app/internal/utils"
)

func main() {
    config.LoadConfig()
    database.Connect()
    cache.Connect()

    db := database.GetDB()
    database.Migrate(db)
    redisClient := cache.GetClient()

    var version string
    err := db.Raw("SELECT version()").Scan(&version).Error
    if err != nil {
        log.Fatalf("Failed to execute query: %v", err)
    }
    fmt.Printf("PostgreSQL version: %s\n", version)

    err = redisClient.Set(cache.Ctx, "key", "Hello, Redis!", 0).Err()
    if err != nil {
        log.Fatalf("Failed to set key in Redis: %v", err)
    }

    val, err := redisClient.Get(cache.Ctx, "key").Result()
    if err != nil {
        log.Fatalf("Failed to get key from Redis: %v", err)
    }
    fmt.Printf("Value from Redis: %s\n", val)

    cron.StartHardDeleteWorker()
    cron.CleanUpExpiredLinks()

	router := gin.Default()

    router.Use(cors.Default()) 
    router.Use(appUtils.UnauthenticatedRateLimiterMiddleware())

	routes.SetupRoutes(router)

    // setup hello world route
    router.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Welcome to File Sharing App API (Tradmarkia Assignment)",
        })
    })

    router.Run(":8080")
}
