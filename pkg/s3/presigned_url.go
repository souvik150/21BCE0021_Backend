package s3

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appConfig "github.com/souvik150/file-sharing-app/internal/config"
)

func GeneratePresignedURL(objectKey string) (string, error) {
	accessKey := appConfig.AppConfig.AWSAccessKey
	secretKey := appConfig.AppConfig.AWSSecretKey
	region := appConfig.AppConfig.AWSRegion
	bucketName := appConfig.AppConfig.BucketName

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion(region),
	)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return "", fmt.Errorf("failed to load AWS config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)
	expiryDuration := 15 * time.Minute

	presignedReq, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(expiryDuration))

	if err != nil {
		log.Printf("Failed to generate presigned URL: %v", err)
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	log.Printf("Generated signed URL: %s", presignedReq.URL)

	return presignedReq.URL, nil
}