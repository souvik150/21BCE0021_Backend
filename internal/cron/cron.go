package cron

import (
	"log"
	"time"

	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
)
func CleanUpExpiredLinks() {
	ticker := time.NewTicker(15 * time.Minute)
	log.Println("Starting link cleanup worker")
	dbClient := database.GetDB()
	go func() {
		for range ticker.C {
			dbClient.Where("expires_at <= ?", time.Now()).Delete(&models.SharedLink{})
		}
	}()
}