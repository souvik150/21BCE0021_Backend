package s3

import (
	"bytes"
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appConfig "github.com/souvik150/file-sharing-app/internal/config"
	"github.com/souvik150/file-sharing-app/pkg/utils"
)

func DownloadFile(fileID string) ([]byte, error) {
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
		return nil, err
	}

	bucket := appConfig.AppConfig.BucketName
	s3Client := s3.NewFromConfig(cfg)
	log.Printf("Downloading file: %s", fileID)
	resp, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileID),
	})
	if err != nil {
		log.Printf("Failed to download file: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	encryptedData := new(bytes.Buffer)
	_, err = io.Copy(encryptedData, resp.Body)
	if err != nil {
		log.Printf("Failed to read encrypted file data: %v", err)
		return nil, err
	}

	decryptedData, err := utils.Decrypt(encryptedData.Bytes(), encryptionKey)
	if err != nil {
		log.Printf("Failed to decrypt file: %v", err)
		return nil, err
	}

	return decryptedData, nil
}
