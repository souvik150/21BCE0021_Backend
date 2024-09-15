package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
)

func DeleteFileHandler(c *gin.Context) {
	fileId := c.Param("id")
	if fileId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	dbClient := database.GetDB()
	var file models.File

	if err := dbClient.Where("id = ?", fileId).First(&file).Error; err != nil {
		log.Printf("Error retrieving file from database: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	file.DeletedAt = time.Now()	

	if err := dbClient.Save(&file).Error; err != nil {
		log.Printf("Error marking file as deleted: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted (soft delete)"})
}

func GetUserDeletedFilesHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("Error getting userID from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get userID from context",
		})
		return
	}

	dbClient := database.GetDB()
	var deletedFiles []models.File

	if err := dbClient.Where("deleted_at IS NOT NULL AND owner_id = ?", userID).Find(&deletedFiles).Error; err != nil {
		log.Printf("Error retrieving user deleted files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get deleted files",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deletedFiles": deletedFiles})
}
