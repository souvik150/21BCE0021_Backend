package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/souvik150/file-sharing-app/internal/cache"
	appConfig "github.com/souvik150/file-sharing-app/internal/config"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/schemas"
	"github.com/souvik150/file-sharing-app/pkg/s3"
)

func GenerateLinkHandler(c *gin.Context) {
	fileId := c.Param("id")

	if fileId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "fileId query parameter is required",
		})
		return
	}

	cacheClient := cache.GetClient()

	cacheData, err := cacheClient.Get(cache.Ctx, fileId).Result()
	if err == nil && cacheData != "" {
		var fileCache schemas.FileCache
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

	link, err := s3.GeneratePresignedURL(s3ObjectName)
	if err != nil {
		log.Printf("Error generating presigned URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate presigned URL",
		})
		return
	}

	fileCache := schemas.FileCache{
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


func ShareFileHandler(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {  
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "File ID is required",
			"error":   "File ID is not provided",
		})
		return
	}

	shareToken := uuid.New().String()
	expiresAt := time.Now().Add(15 * time.Minute)

	dbClient := database.GetDB()

	var file models.File
	if err := dbClient.Where("id = ?", fileID).First(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File not found in database"})
		return
	}

	sharedLink := models.SharedLink{
		FileID:     uuid.MustParse(fileID),
		FileName:  file.FileName,
		ShareToken: shareToken,
		ExpiresAt:  expiresAt,
	}

	if err := dbClient.Create(&sharedLink).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate shareable link"})
		return
	}

	backendURL := appConfig.AppConfig.BackendURL
	shareableLink := fmt.Sprintf("%s/share/%s",backendURL , shareToken)

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"message": "Shareable link generated successfully. Expires in 15 minutes",
		"link": shareableLink,
	})
}