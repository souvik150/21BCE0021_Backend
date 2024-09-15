package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/souvik150/file-sharing-app/internal/handlers"
	"github.com/souvik150/file-sharing-app/internal/middleware"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/register", handlers.RegisterUserHandler)
	r.POST("/login", handlers.LoginUserHandler)
	r.GET("/share/:share_token", handlers.ServeSharedFileHandler)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware()) 
	{
		protected.POST("/upload", handlers.UploadMultipleFilesHandler)
		protected.GET("/generate/:id", handlers.ShareFileHandler)
		protected.DELETE("/delete/:id", handlers.DeleteFileHandler)
		protected.GET("/deleted-files", handlers.GetUserDeletedFilesHandler)
		protected.GET("/my-files", handlers.GetUserFilesHandler)
		protected.PATCH("/update", handlers.UpdateFileHandler)
		protected.GET("/me", handlers.GetCurrentUser)
	}
}
