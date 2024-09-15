package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/cache"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/schemas"
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file in database"})
			return
	}

	// delete from cache
	cacheClient := cache.GetClient()
	if err := cacheClient.Del(cache.Ctx, fileId).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file from cache"})
			return
	}

	fileResponse := schemas.RenameFileResponse{
			ID:        file.ID.String(),
			FileName:  file.FileName,
			CreatedAt: file.CreatedAt,
			FileType: file.FileType,
			UpdatedAt: file.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "File renamed successfully",
			"data":    fileResponse,
	})
}