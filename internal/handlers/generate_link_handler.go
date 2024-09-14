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
	"github.com/souvik150/file-sharing-app/internal/s3service"
)

func GenerateLinkHandler(c *gin.Context) {
    fileName := c.Query("id")
    bucket := "trademarkia-assignment"

    if fileName == "" || bucket == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "fileName and bucket query parameters are required",
        })
        return
    }

    client := cache.GetClient()

    // cache hit
    cacheData, err := client.Get(cache.Ctx, fileName).Result()
    if err == nil {
        returnCachedURL(c, cacheData, fileName)
        return
    }

    if err.Error() != "redis: nil" {
        log.Printf("Error getting file from cache: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to get file from cache",
        })
        return
    }

    // cache miss
    link, err := s3service.GeneratePresignedURL(fileName)
    if err != nil {
        log.Printf("Error generating link: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to generate presigned URL",
        })
        return
    }

    dbClient := database.GetDB()
    var file models.File
    if err := dbClient.Where("file_name = ?", fileName).First(&file).Error; err != nil {
        log.Printf("Error getting file from database: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to get file from database",
        })
        return
    }

    cacheData = prepareCacheData(file, link)
    if err := client.Set(cache.Ctx, file.FileName, cacheData, 5*time.Minute).Err(); err != nil {
        log.Printf("Error caching file: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to cache file",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{"link": link})
}

func returnCachedURL(c *gin.Context, cacheData string, fileName string) {
    var fileData map[string]interface{}
    if err := json.Unmarshal([]byte(cacheData), &fileData); err != nil {
        log.Printf("Error unmarshalling cache data: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to get file from cache",
        })
        return
    }

    log.Printf("Returning URL from cache for file: %s", fileName)
    c.JSON(http.StatusOK, gin.H{
        "link": fileData["url"],
    })
}

func prepareCacheData(file models.File, link string) string {
    fileCache := models.File{
        ID:         uuid.New(),
        FileName:   file.FileName,
        OwnerID:    file.OwnerID,
        Size:       file.Size,
        AccessedAt: time.Now(),
    }
    cacheData := map[string]interface{}{
        "file": fileCache,
        "url":  link,
    }
    cacheDataBytes, _ := json.Marshal(cacheData)
    return string(cacheDataBytes)
}
