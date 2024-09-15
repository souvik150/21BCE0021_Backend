package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/s3service"
)

func ServeSharedFileHandler(c *gin.Context) {
	shareToken := c.Param("share_token")
	if shareToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "share_token is required"})
		return
	}

	dbClient := database.GetDB()
	var sharedLink models.SharedLink

	// Look up the shared link in the database
	if err := dbClient.Where("share_token = ?", shareToken).First(&sharedLink).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	// Check if the link has expired
	if time.Now().After(sharedLink.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Link has expired"})
		return
	}

	fileData, err := s3service.DownloadFile(sharedLink.FileID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+sharedLink.FileName)
	c.Data(http.StatusOK, "application/octet-stream", fileData)
}
