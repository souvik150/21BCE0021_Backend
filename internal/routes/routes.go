package routes

import (
	"github.com/gin-gonic/gin"

	fileHandlers "github.com/souvik150/file-sharing-app/internal/handlers/file"
	userHandlers "github.com/souvik150/file-sharing-app/internal/handlers/user"
	"github.com/souvik150/file-sharing-app/pkg/middleware"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/register", userHandlers.RegisterUserHandler)
	r.POST("/login", userHandlers.LoginUserHandler)
	r.GET("/share/:share_token", fileHandlers.ServeSharedFileHandler)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware()) 
	{
		protected.POST("/upload", fileHandlers.UploadMultipleFilesHandler)
		protected.GET("/generate/:id", fileHandlers.ShareFileHandler)
		protected.DELETE("/delete/:id", fileHandlers.DeleteFileHandler)
		protected.GET("/deleted-files", fileHandlers.GetUserDeletedFilesHandler)
		protected.GET("/my-files", fileHandlers.GetUserFilesHandler)
		protected.PATCH("/update", fileHandlers.UpdateFileHandler)
		protected.GET("/me", userHandlers.GetCurrentUser)
	}
}
