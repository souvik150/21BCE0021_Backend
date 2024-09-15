package s3

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	appConfig "github.com/souvik150/file-sharing-app/internal/config"
	"github.com/souvik150/file-sharing-app/internal/socket"
)

func ProcessFilesAsync(fileNames []string, ownerID string) {
	var wg sync.WaitGroup
	wg.Add(len(fileNames))

	for _, fileName := range fileNames {
		go func(fileName string) {
			defer wg.Done()
			filePath := filepath.Join(fileName)

			file, err := os.Open(filePath)
			if err != nil {
				log.Printf("❌ Error opening file %s: %v", fileName, err)
				return
			}
			defer file.Close()

			fileInfo, err := file.Stat()
			if err != nil {
				log.Printf("❌ Error getting file info for %s: %v", fileName, err)
				return
			}
			
			fileName = fileName[7:]
			bucket := appConfig.AppConfig.BucketName
			err = UploadFileConcurrently(bucket, fileName, file, fileInfo.Size())
			if err != nil {
				log.Printf("❌ Error uploading file %s to S3: %v", fileName, err)
				return
			}

			log.Printf("✅ File %s successfully uploaded to S3", fileName)
		}(fileName)
	}

	wg.Wait()

	for _, fileName := range fileNames {
		err := os.Remove(fileName)
			if err != nil {
				log.Printf("⚠️ Error deleting local file %s after upload: %v", fileName, err)
			} else {
				log.Printf("🗑️ Local file %s deleted after successful upload", fileName)
			}
		}	
	
	socket.NotifyUser(ownerID, "All files processed and uploaded to S3.")
	log.Println("All files processed and uploaded to S3.")
}