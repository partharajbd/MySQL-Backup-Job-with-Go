package storage

import (
	"context"
	"fmt"
	"mysql-database-backup-manager/configs"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadToS3(filePath string, appConfig *configs.AppConfig, key string) error {
	fmt.Println("> Uploading database: ", filepath.Base(filePath))
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(appConfig.Region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint := appConfig.S3Endpoint; endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	s3Key := key
	if appConfig.S3RootDir != "" {
		s3Key = filepath.Join(appConfig.S3RootDir, key)
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(appConfig.S3Bucket),
		Key:    aws.String(s3Key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}
