package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
)

func SearchFilesHandler(c *gin.Context) {
	ownerID := c.GetString("userID")
	fileName := c.Query("fileName")
	fileType := c.Query("fileType")
	uploadDate := c.Query("uploadDate")

	var parsedDate time.Time
	if uploadDate != "" {
			var err error
			parsedDate, err = time.Parse("2006-01-02", uploadDate)
			if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
							"error": "Invalid date format. Expected YYYY-MM-DD.",
					})
					return
			}
	}

	var files []models.File
	dbClient := database.GetDB()

	query := dbClient.Where("owner_id = ?", ownerID)
	if fileName != "" {
			query = query.Where("file_name ILIKE ?", "%"+fileName+"%")
	}
	if fileType != "" {
			query = query.Where("file_type = ?", fileType)
	}
	if !parsedDate.IsZero() {
			query = query.Where("DATE(created_at) = ?", parsedDate)
	}

	if err := query.Find(&files).Error; err != nil {
			log.Printf("Error searching files: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to retrieve files",
			})
			return
	}

	c.JSON(http.StatusOK, gin.H{
			"files": files,
	})
}
