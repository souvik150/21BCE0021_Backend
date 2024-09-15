package handlers

import (
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	appConfig "github.com/souvik150/file-sharing-app/internal/config"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/schemas"
)

func LoginUserHandler(c *gin.Context) {
	var input schemas.LoginInput
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"message": "Invalid request payload",
			"error": err.Error(),
		})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": false,
			"message": "Invalid email or password",
			"error": err.Error(),
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": false,
			"message": "Invalid email or password",
			"error": err.Error(),
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	
	var jwtSecretKey = []byte(appConfig.AppConfig.EncryptionKey)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": false,
			"message": "Failed to generate token",
			"error": err.Error(),
		})
		return
	}

	var userResponse schemas.LoginResponse
	userResponse.Email = user.Email
	userResponse.Token = tokenString

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"message": "User logged in successfully",
		"data":  userResponse,
	})
}
