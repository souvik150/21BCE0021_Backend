package database

import (
	"log"

	"gorm.io/gorm"

	"github.com/souvik150/file-sharing-app/internal/models"
)

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.File{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migration completed successfully.")
}