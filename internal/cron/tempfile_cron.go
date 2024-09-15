package cron

import (
	"log"
	"time"

	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/s3service"
)

func StartHardDeleteWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	log.Println("Starting hard delete worker")
	go func() {
		for range ticker.C {
			hardDeleteExpiredFiles()
		}
	}()
}

func CleanUpExpiredLinks() {
	ticker := time.NewTicker(5 * time.Minute)
	log.Println("Starting link cleanup worker")
	dbClient := database.GetDB()
	go func() {
		for range ticker.C {
			dbClient.Where("expires_at <= ?", time.Now()).Delete(&models.SharedLink{})
		}
	}()
}

func hardDeleteExpiredFiles() {
	dbClient := database.GetDB()
	var expiredFiles []models.File
	cutoffDate := time.Now().Add(-2 * time.Minute)

	if err := dbClient.Where("deleted_at IS NOT NULL AND deleted_at <= ?", cutoffDate).Find(&expiredFiles).Error; err != nil {
		log.Printf("Error retrieving expired files for hard delete: %v", err)
		return
	}

	for _, file := range expiredFiles {
		if err := s3service.DeleteFileFromS3(file.ID.String()); err != nil {
			log.Printf("Error deleting file from S3: %v", err)
			continue
		}

		if err := dbClient.Delete(&file).Error; err != nil {
			log.Printf("Error deleting file metadata from database: %v", err)
		}

		log.Printf("File %s deleted permanently", file.FileName)
	}
}