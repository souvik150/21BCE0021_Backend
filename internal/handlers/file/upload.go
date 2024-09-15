package handlers

import (
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/souvik150/file-sharing-app/internal/database"
	"github.com/souvik150/file-sharing-app/internal/models"
	"github.com/souvik150/file-sharing-app/internal/schemas"
	"github.com/souvik150/file-sharing-app/pkg/s3"
)

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

	var uploadedFiles []string
	var uploadFilesPaths []string
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
				DeletedStatus: false,
				DeletedAt: time.Time{},
			}

			err = db.Create(&newFile).Error
			if err != nil {
				log.Printf("Error creating file in database: %v", err)
				return
			}

			tmpDir := "/local"
			if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
				log.Printf("Creating /local directory...")
				err = os.Mkdir(tmpDir, 0755)
				if err != nil {
					log.Printf("Error creating /local directory: %v", err)
					return
				}
			}

			filePath := filepath.Join(tmpDir, newFile.ID.String())
			log.Printf("Saving file locally at path: %s", filePath)

			out, err := os.Create(filePath)
			if err != nil {
				log.Printf("Error creating local file: %v", err)
				return
			}
			defer out.Close()

			_, err = io.Copy(out, file)
			if err != nil {
				log.Printf("Error writing file to local storage: %v", err)
				return
			}

			mu.Lock()
			uploadedFiles = append(uploadedFiles, newFile.FileName)
			uploadFilesPaths = append(uploadFilesPaths, filePath)
			mu.Unlock()

			log.Printf("File %s stored locally at %s", header.Filename, filePath)
		}(header)
	}

	wg.Wait()

	go s3.ProcessFilesAsync(uploadFilesPaths, parsedUserID.String())

	var res schemas.UploadedFileResponse
	res.FileNames = uploadedFiles

	if len(uploadedFiles) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Files uploaded. Undergoing processing",
			"data": res,
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to upload files",
			"error": "Failed to store files locally",
		})
	}
}
