package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appConfig "github.com/souvik150/file-sharing-app/internal/config"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
)

func ShareFileHandler(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {  
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id is required"})
		return
	}

	shareToken := uuid.New().String()
	expiresAt := time.Now().Add(15 * time.Minute)

	dbClient := database.GetDB()

	// get the file from the database
	var file models.File
	if err := dbClient.Where("id = ?", fileID).First(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file from database"})
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

	c.JSON(http.StatusOK, gin.H{"shareable_link": shareableLink})
}