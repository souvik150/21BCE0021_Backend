package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
)

type FileResponse struct {
	ID          uuid.UUID `json:"id"`
	FileName    string    `json:"file_name"`
	Size        int64     `json:"size"`
	FileType    string    `json:"file_type"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	AccessedAt  string    `json:"accessed_at"`
	DeletedStatus bool    `json:"deleted_status"`
}

type UserResponse struct {
	ID        uuid.UUID      `json:"id"`
	Email     string         `json:"email"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
	Files     []FileResponse `json:"files"`
}

func GetUserFilesHandler(c *gin.Context) {
	
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("Error getting userID from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get userID from context",
		})
		return
	}

	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		log.Printf("Error parsing userID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid userID format",
		})
		return
	}

	db := database.GetDB()
	var user models.User

	if err := db.Preload("Files").Where("id = ?", parsedUserID).First(&user).Error; err != nil {
		log.Printf("Error fetching user and files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user and files",
		})
		return
	}

	userResponse := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	for _, file := range user.Files {
		fileResponse := FileResponse{
			ID:           file.ID,
			FileName:     file.FileName,
			Size:         file.Size,
			FileType:     file.FileType,
			CreatedAt:    file.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:    file.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			AccessedAt:   file.AccessedAt.Format("2006-01-02T15:04:05Z"),
			DeletedStatus: file.DeletedStatus,
		}
		userResponse.Files = append(userResponse.Files, fileResponse)
	}

	c.JSON(http.StatusOK, userResponse)
}
