package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/schemas"
)

func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("Error getting userID from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to get userID from context",
			"error":   "userID not found in context",
		})
		return
	}

	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		log.Printf("Error parsing userID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invalid userID format",
			"error":   err.Error(),
		})
		return
	}

	db := database.GetDB()
	var user models.User

	if err := db.Where("id = ?", parsedUserID).First(&user).Error; err != nil {
		log.Printf("Error fetching user from database: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "User not found",
			"error":   err.Error(),
		})
		return
	}

	userResponse := schemas.GetCurrentUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User fetched successfully",
		"data":    userResponse,
	})
}
