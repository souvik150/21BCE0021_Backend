package cache

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"

	"github.com/souvik150/file-sharing-app/internal/config"
)

var (
    Ctx         = context.Background()
    RedisClient *redis.Client
)

func Connect() {
    redisURI := config.AppConfig.RedisURI

    options, err := redis.ParseURL(redisURI)
    if err != nil {
        log.Fatalf("Failed to parse Redis URI: %v", err)
    }

    RedisClient = redis.NewClient(options)

    _, err = RedisClient.Ping(Ctx).Result()
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }

    fmt.Println("✅ Connected to Redis successfully.")
}

func GetClient() *redis.Client {
    return RedisClient
}
