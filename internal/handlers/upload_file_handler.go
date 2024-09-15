package handlers

import (
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/souvik150/file-sharing-app/internal/cache"
	appConfig "github.com/souvik150/file-sharing-app/internal/config"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/s3service"
)

type FileCache struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	FileName  string    `json:"file_name"`
	Size      int64     `json:"size"`
	FileType  string    `json:"file_type"`
}

func UploadMultipleFilesHandler(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get form from request",
		})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No files were uploaded",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get userID from context",
		})
		return
	}

	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse userID",
		})
		return
	}

	db := database.GetDB()
	redisClient := cache.GetClient()

	var uploadedFiles []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(files))

	for _, header := range files {
		go func(header *multipart.FileHeader) {
			defer wg.Done()

			file, err := header.Open()
			if err != nil {
				log.Printf("Error opening file: %v", err)
				return
			}
			defer file.Close()

			fileExt := filepath.Ext(header.Filename)
			if len(fileExt) > 0 {
				fileExt = fileExt[1:]
			} else {
				log.Printf("File %s has no extension", header.Filename)
			}

			newFile := models.File{
				ID:         uuid.New(),
				FileName:   header.Filename,
				OwnerID:    parsedUserID,
				Size:       header.Size,
				FileType:   fileExt,
				CreatedAt:  time.Now(),
				AccessedAt: time.Now(),
				UpdatedAt:  time.Now(),
			}

			err = db.Create(&newFile).Error
			if err != nil {
				log.Printf("Error creating file in database: %v", err)
				return
			}

			fileURL, err := s3service.GeneratePresignedURL(newFile.ID.String())
			if err != nil {
				log.Printf("Error generating URL for file: %v", err)
				return
			}

			cacheData := FileCache{
				ID:        newFile.ID,
				URL:       fileURL,
				FileName:  newFile.FileName,
				Size:      newFile.Size,
				FileType:  newFile.FileType,
			}
			
			cacheDataBytes, err := json.Marshal(cacheData)
			if err != nil {
				log.Printf("Error marshalling cache data: %v", err)
				return
			}

			err = redisClient.Set(cache.Ctx, newFile.ID.String(), cacheDataBytes, 5*time.Minute).Err()
			if err != nil {
				log.Printf("Error caching file: %v", err)
				return
			}

			bucket := appConfig.AppConfig.BucketName
			uploadedFileName := newFile.ID.String()
			log.Printf("Uploading file: %s to bucket: %s with ID: %s", header.Filename, bucket, uploadedFileName)

			err = s3service.UploadFileConcurrently(bucket, uploadedFileName, file, header.Size)
			if err != nil {
				log.Printf("Error uploading file: %v", err)
				return
			}

			mu.Lock()
			uploadedFiles = append(uploadedFiles, uploadedFileName)
			mu.Unlock()

		}(header)
	}

	wg.Wait()

	if len(uploadedFiles) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":       "Files uploaded successfully",
			"uploadedFiles": uploadedFiles,
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No files were uploaded successfully",
		})
	}
}
