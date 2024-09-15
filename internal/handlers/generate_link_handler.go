package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/cache"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/s3service"
)

func GenerateLinkHandler(c *gin.Context) {
	fileId := c.Query("id")

	if fileId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "fileId query parameter is required",
		})
		return
	}

	cacheClient := cache.GetClient()

	cacheData, err := cacheClient.Get(cache.Ctx, fileId).Result()
	if err == nil && cacheData != "" {
		var fileCache FileCache
		if err := json.Unmarshal([]byte(cacheData), &fileCache); err != nil {
			log.Printf("Error unmarshalling cache data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to unmarshal cache data",
			})
			return
		}

		log.Printf("Returning cached link for file: %s", fileCache.FileName)
		c.JSON(http.StatusOK, gin.H{"link": fileCache.URL})
		return
	}

	dbClient := database.GetDB()
	var fileDb models.File

	if err := dbClient.Where("id = ?", fileId).First(&fileDb).Error; err != nil {
		log.Printf("Error getting file from database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get file from database",
		})
		return
	}

	log.Printf("Generating link for file: %s", fileDb.FileName)

	s3ObjectName := fileDb.ID.String() + "." + fileDb.FileType

	link, err := s3service.GeneratePresignedURL(s3ObjectName)
	if err != nil {
		log.Printf("Error generating presigned URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate presigned URL",
		})
		return
	}

	fileCache := FileCache{
		ID:       fileDb.ID,
		URL:      link,
		FileName: fileDb.FileName,
		Size:     fileDb.Size,
		FileType: fileDb.FileType,
	}

	cacheDataBytes, err := json.Marshal(fileCache)
	if err != nil {
		log.Printf("Error marshalling file cache: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cache data",
		})
		return
	}

	err = cacheClient.Set(cache.Ctx, fileId, cacheDataBytes, 15*time.Minute).Err()
	if err != nil {
		log.Printf("Error setting cache: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to set cache",
		})
		return
	}

	log.Printf("Returning generated link for file: %s", fileDb.FileName)
	c.JSON(http.StatusOK, gin.H{"link": link})
}
