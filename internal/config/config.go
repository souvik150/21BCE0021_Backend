package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
    PostgresURI string
    RedisURI    string
		AWSAccessKey string
		AWSSecretKey string
		AWSRegion string
		EncryptionKey string
		BucketName string
		BackendURL string
}

var AppConfig *Config

func LoadConfig() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found or it could not be loaded. Proceeding with system environment variables.")
    }

    viper.AutomaticEnv()

    postgresURI := viper.GetString("POSTGRES_URI")
    if postgresURI == "" {
        log.Fatal("POSTGRES_URI is required")
    }

    redisURI := viper.GetString("REDIS_URI")
    if redisURI == "" {
        log.Fatal("REDIS_URI is required")
    }

		accessKey := viper.GetString("AWS_ACCESS_KEY_ID")
		if accessKey == "" {
			log.Fatal("AWS_ACCESS_KEY_ID is required")
		}

		secretKey := viper.GetString("AWS_SECRET_ACCESS_KEY")
		if secretKey == "" {
			log.Fatal("AWS_SECRET_ACCESS_KEY is required")
		}

		region := viper.GetString("AWS_REGION")
		if region == "" {
			log.Fatal("AWS_REGION is required")
		}

		encryptionKey := viper.GetString("ENCRYPTION_KEY")
		if encryptionKey == "" {
			log.Fatal("ENCRYPTION_KEY is required")
		}

		bucketName := viper.GetString("AWS_BUCKET_NAME")
		if bucketName == "" {
			log.Fatal("AWS_BUCKET_NAME is required")
		}

		backendURL := viper.GetString("BACKEND_URL")
		if backendURL == "" {
			log.Fatal("BACKEND_URL is required")
		}

    AppConfig = &Config{
        PostgresURI: postgresURI,
        RedisURI:    redisURI,
				AWSAccessKey: accessKey,
				AWSSecretKey: secretKey,
				AWSRegion: region,
				EncryptionKey: encryptionKey,
				BucketName: bucketName,
				BackendURL: backendURL,
    }
}
