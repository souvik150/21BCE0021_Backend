package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/handlers"
	"github.com/souvik150/file-sharing-app/internal/middleware"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/register", handlers.RegisterUserHandler)
	r.POST("/login", handlers.LoginUserHandler)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware()) 
	{
		protected.POST("/upload", handlers.UploadMultipleFilesHandler)
		protected.GET("/share/:id", handlers.GenerateLinkHandler)
		protected.GET("/rename", handlers.UpdateFileHandler)
		protected.GET("/my-files", handlers.GetUserFilesHandler)
		protected.GET("/me", handlers.GetCurrentUser)
	}
}
