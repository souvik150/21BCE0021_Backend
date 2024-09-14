package handlers

import (
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/souvik150/file-sharing-app/internal/cache"
	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/s3service"
)

func UploadMultipleFilesHandler(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error getting multipart form: %v", err)
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
		log.Printf("Error getting userID from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get userID from context",
		})
		return
	}
	userID, err = uuid.Parse(userID.(string))
	if err != nil {
		log.Printf("Error parsing userID: %v", err)
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

			newFile := models.File{
				ID:         uuid.New(),
				FileName:   header.Filename,
				OwnerID:    userID.(uuid.UUID),
				Size:       header.Size,
				FileType:   strings.Split(header.Filename, ".")[1],
				CreatedAt:  time.Now(),
				AccessedAt: time.Now(),
				UpdatedAt:  time.Now(),
			}

			err = db.Create(&newFile).Error
			if err != nil {
				log.Printf("Error creating file in database: %v", err)
				return
			}

			fileURL, err := s3service.GeneratePresignedURL(newFile.FileName)
			if err != nil {
				log.Printf("Error generating URL for file: %v", err)
				return
			}

			cacheData := map[string]interface{}{
				"file": newFile,
				"url":  fileURL,
			}
			cacheDataBytes, err := json.Marshal(cacheData)
			if err != nil {
				log.Printf("Error marshalling cache data: %v", err)
				return
			}

			err = redisClient.Set(cache.Ctx, header.Filename, cacheDataBytes, 5*time.Minute).Err()
			if err != nil {
				log.Printf("Error caching file: %v", err)
				return
			}

			bucket := "trademarkia-assignment"
			log.Printf("Uploading file: %s to bucket: %s", header.Filename, bucket)

			err = s3service.UploadFileConcurrently(bucket, header.Filename, file, header.Size)
			if err != nil {
				log.Printf("Error uploading file: %v", err)
				return
			}

			mu.Lock()
			uploadedFiles = append(uploadedFiles, header.Filename)
			mu.Unlock()

		}(header) 
	}

	wg.Wait()

	if len(uploadedFiles) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":        "Files uploaded successfully",
			"uploadedFiles":  uploadedFiles,
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No files were uploaded successfully",
		})
	}
}
