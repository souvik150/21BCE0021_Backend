package s3

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appConfig "github.com/souvik150/file-sharing-app/internal/config"
)

func DeleteFileFromS3(fileKey string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return err
	}

	s3Client := s3.NewFromConfig(cfg)

	_, err = s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(appConfig.AppConfig.BucketName),
		Key:    aws.String(fileKey),
	})

	if err != nil {
		log.Printf("Error deleting file from S3: %v", err)
		return err
	}

	log.Printf("File deleted from S3: %s", fileKey)
	return nil
}