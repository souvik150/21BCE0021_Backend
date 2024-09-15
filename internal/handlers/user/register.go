package handlers

import (
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/schemas"
)

func RegisterUserHandler(c *gin.Context) {
	var input schemas.RegisterUserInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to hash password",
			"error":   err.Error(),
		})
		return
	}

	user := models.User{
		Email:    input.Email,
		Password: string(hashedPassword),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to create user",
			"error":   err.Error(),
		})
		return
	}

	var userResponse schemas.UserSignupResponse
	userResponse.ID = user.ID
	userResponse.Email = user.Email

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "User created successfully. Please login to continue",
		"data":    userResponse,
	})
}
