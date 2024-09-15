package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	appConfig "github.com/souvik150/file-sharing-app/internal/config"
	"github.com/souvik150/file-sharing-app/pkg/utils"
)

const (
	partSize         = 5 * 1024 * 1024
)

func UploadFileConcurrently(bucket, key string, file multipart.File, fileSize int64) error {
	accessKey := appConfig.AppConfig.AWSAccessKey
	secretKey := appConfig.AppConfig.AWSSecretKey
	region := appConfig.AppConfig.AWSRegion
	encryptionKey := []byte(appConfig.AppConfig.EncryptionKey)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	s3Svc := s3.NewFromConfig(cfg)

	if fileSize <= partSize {
		log.Println("File is small, performing single upload")
		uploadBuffer := new(bytes.Buffer)
		_, err := io.Copy(uploadBuffer, file)
		if err != nil {
			log.Printf("Failed to read small file: %v", err)
			return fmt.Errorf("failed to read file: %v", err)
		}

		// Encrypt the file content before uploading
		encryptedData, err := utils.Encrypt(uploadBuffer.Bytes(), encryptionKey)
		if err != nil {
			log.Printf("Failed to encrypt file: %v", err)
			return fmt.Errorf("failed to encrypt file: %v", err)
		}

		_, err = s3Svc.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   bytes.NewReader(encryptedData),
		})
		if err != nil {
			log.Printf("Failed to upload file as single part: %v", err)
			return fmt.Errorf("failed to upload file: %v", err)
		}

		log.Println("File uploaded successfully using single upload")
		return nil
	}

	createResp, err := s3Svc.CreateMultipartUpload(context.TODO(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Failed to initiate multipart upload: %v", err)
		return fmt.Errorf("failed to initiate multipart upload: %v", err)
	}

	uploadID := createResp.UploadId
	var completedParts []types.CompletedPart
	var mu sync.Mutex
	
	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	partNum := 1

	for {
		buffer := make([]byte, partSize)
		bytesRead, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading file: %v", err)
			return fmt.Errorf("error reading file: %v", err)
		}

		wg.Add(1)
		go func(partNum int, buffer []byte) {
			defer wg.Done()
			uploadResp, err := s3Svc.UploadPart(context.TODO(), &s3.UploadPartInput{
				Bucket:     aws.String(bucket),
				Key:        aws.String(key),
				PartNumber: aws.Int32(int32(partNum)),
				UploadId:   uploadID,
				Body:       bytes.NewReader(buffer[:bytesRead]),
			})
			if err != nil {
				log.Printf("Failed to upload part %d: %v", partNum, err)
				select {
				case errCh <- fmt.Errorf("failed to upload part %d: %v", partNum, err):
				default:
				}
				return
			}

			mu.Lock()
			completedParts = append(completedParts, types.CompletedPart{
				ETag:       uploadResp.ETag,
				PartNumber: aws.Int32(int32(partNum)),
			})
			mu.Unlock()
		}(partNum, buffer[:bytesRead])
		partNum++
	}

	wg.Wait()

	select {
	case err := <-errCh:
		log.Printf("Aborting multipart upload due to error: %v", err)
		_, abortErr := s3Svc.AbortMultipartUpload(context.TODO(), &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(bucket),
			Key:      aws.String(key),
			UploadId: uploadID,
		})
		if abortErr != nil {
			log.Printf("Failed to abort multipart upload: %v", abortErr)
		}
		return err
	default:
	}

	sort.Slice(completedParts, func(i, j int) bool {
		return *completedParts[i].PartNumber < *completedParts[j].PartNumber
	})

	log.Println("Completed Parts:")
	for _, part := range completedParts {
		log.Printf("Part Number: %d, ETag: %s", *part.PartNumber, *part.ETag)
	}

	_, err = s3Svc.CompleteMultipartUpload(context.TODO(), &s3.CompleteMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		UploadId: uploadID,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		log.Printf("Failed to complete multipart upload: %v", err)
		return fmt.Errorf("failed to complete multipart upload: %v", err)
	}

	log.Println("File uploaded successfully using multipart upload")
	return nil
}