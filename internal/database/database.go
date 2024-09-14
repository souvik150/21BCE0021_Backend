package database

import (
	"log"
	"net/url"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/souvik150/file-sharing-app/internal/config"
)

var DB *gorm.DB

func Connect() {
	serviceURI := config.AppConfig.PostgresURI

	connURL, err := url.Parse(serviceURI)
	if err != nil {
		log.Fatalf("Invalid database URL: %v", err)
	}

	q := connURL.Query()
	q.Set("sslmode", "verify-ca")
	q.Set("sslrootcert", "ca.pem")
	connURL.RawQuery = q.Encode()

	dsn := connURL.String()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	DB = db

	err = db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		log.Fatalf("Failed to enable uuid-ossp extension: %v", err)
	}
	log.Println("uuid-ossp extension is enabled.")

	
	log.Println("Connected to PostgreSQL successfully.")
}

func GetDB() *gorm.DB {
	return DB
}
