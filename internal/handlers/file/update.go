package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/cache"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
)

func UpdateFileHandler(c *gin.Context) {
	fileId := c.Query("fileId")
	newFileName := c.Query("newFileName")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userID from context"})
		return
	}

	if fileId == "" || newFileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fileId and newFileName query parameters are required"})
		return
	}

	dbClient := database.GetDB()
	var file models.File

	if err := dbClient.Where("id = ?", fileId).First(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File not found in database"})
		return
	}

	if file.OwnerID.String() != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to rename this file"})
		return
	}

	file.FileName = newFileName
	file.UpdatedAt = time.Now()

	if err := dbClient.Save(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file"})
		return
	}

	client := cache.GetClient()
	cacheData, err := client.Get(cache.Ctx, fileId).Result()
	if err == nil {
		if err := renameFileInCache(cacheData, fileId, newFileName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rename file in cache"})
			return
		}
	} else if err.Error() != "redis: nil" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file from cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File renamed successfully", "file": file})
}

func renameFileInCache(cacheData string, fileId, newFileName string) error {
	client := cache.GetClient()
	var fileData map[string]interface{}
	if err := json.Unmarshal([]byte(cacheData), &fileData); err != nil {
		return err
	}

	fileData["file"].(map[string]interface{})["FileName"] = newFileName

	updatedCacheDataBytes, err := json.Marshal(fileData)
	if err != nil {
		return err
	}

	if err := client.Del(cache.Ctx, fileId).Err(); err != nil {
		return err
	}

	if err := client.Set(cache.Ctx, newFileName, updatedCacheDataBytes, 5*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}
