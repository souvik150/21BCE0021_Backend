package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/souvik150/file-sharing-app/internal/cache"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/schemas"
	"github.com/souvik150/file-sharing-app/pkg/s3"
)

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

	// Get search filters
	fileName := c.Query("name")
	fileType := c.Query("type")	
	uploadDate := c.Query("uploadDate")

	log.Printf("Searching files for user %s with filters: name=%s, type=%s, uploadDate=%s", parsedUserID.String(), fileName, fileType, uploadDate)

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

	db := database.GetDB()
	var files []models.File

	query := db.Where("owner_id = ?", parsedUserID)

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
		log.Printf("Error fetching files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve files",
		})
		return
	}

	userResponse := schemas.FilesResponse{}

	redisClient := cache.GetClient()
	for _, file := range files {
		cacheKey := file.ID.String()

		cachedData, err := redisClient.Get(cache.Ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var cachedFile schemas.FileCache
			if err := json.Unmarshal([]byte(cachedData), &cachedFile); err == nil {
				fileResponse := schemas.FileResponse{
					ID:            cachedFile.ID,
					FileName:      cachedFile.FileName,
					Size:          cachedFile.Size,
					FileType:      cachedFile.FileType,
					CreatedAt:     file.CreatedAt.Format("2006-01-02T15:04:05Z"),
					UpdatedAt:     file.UpdatedAt.Format("2006-01-02T15:04:05Z"),
					AccessedAt:    file.AccessedAt.Format("2006-01-02T15:04:05Z"),
					DeletedStatus: file.DeletedStatus,
				}
				userResponse.Files = append(userResponse.Files, fileResponse)
				log.Printf("File %s (ID: %s) loaded from cache", file.FileName, file.ID.String())
				continue
			}
		}

		s3ObjectName := file.ID.String() + "." + file.FileType
		link, err := s3.GeneratePresignedURL(s3ObjectName)
		if err != nil {
			log.Printf("Error generating presigned URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate presigned URL",
			})
			return
		}

		fileResponse := schemas.FileResponse{
			ID:            file.ID,
			FileName:      file.FileName,
			Size:          file.Size,
			FileType:      file.FileType,
			CreatedAt:     file.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     file.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			AccessedAt:    file.AccessedAt.Format("2006-01-02T15:04:05Z"),
			DeletedStatus: file.DeletedStatus,
		}
		userResponse.Files = append(userResponse.Files, fileResponse)

		fileCache := schemas.FileCache{
			ID:       file.ID,
			URL:      link,
			FileName: file.FileName,
			Size:     file.Size,
			FileType: file.FileType,
		}
		cacheDataBytes, err := json.Marshal(fileCache)
		if err == nil {
			err = redisClient.Set(cache.Ctx, cacheKey, cacheDataBytes, 15*time.Minute).Err()
			if err == nil {
				log.Printf("File %s (ID: %s) cached successfully", file.FileName, file.ID.String())
			} else {
				log.Printf("Error caching file %s (ID: %s): %v", file.FileName, file.ID.String(), err)
			}
		} else {
			log.Printf("Error marshalling file data for cache: %v", err)
		}
		log.Printf("File %s (ID: %s) loaded from database", file.FileName, file.ID.String())
	}

	c.JSON(http.StatusOK , gin.H{
		"success": true,
		"message": "Files fetched successfully",
		"data": userResponse.Files,
	})
}
