package storage

import (
	"context"
	"fmt"
	"mysql-database-backup-manager/configs"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

func DeleteOldBackups(appConfig *configs.AppConfig, retentionDays int) (int, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(appConfig.Region))
	if err != nil {
		return 0, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint := appConfig.S3Endpoint; endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	fmt.Printf("\n=== Cleaning Old Backups (older than %d days) ===\n", retentionDays)

	prefix := ""
	if appConfig.S3RootDir != "" {
		prefix = appConfig.S3RootDir + "/"
	}

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(appConfig.S3Bucket),
		Prefix: aws.String(prefix),
	})

	deletedCount := 0
	var objectsToDelete []types.ObjectIdentifier

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.LastModified.Before(cutoffDate) {
				objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
					Key: obj.Key,
				})
				fmt.Printf("> Marking for deletion: %s (Last Modified: %s)\n",
					*obj.Key, obj.LastModified.Format("2006-01-02 15:04:05"))
			}
		}
	}

	if len(objectsToDelete) == 0 {
		fmt.Println("No old backups found to delete.")
		return 0, nil
	}

	batchSize := 1000
	for i := 0; i < len(objectsToDelete); i += batchSize {
		end := i + batchSize
		if end > len(objectsToDelete) {
			end = len(objectsToDelete)
		}

		batch := objectsToDelete[i:end]
		_, err := client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(appConfig.S3Bucket),
			Delete: &types.Delete{
				Objects: batch,
				Quiet:   aws.Bool(false),
			},
		})
		if err != nil {
			return 0, fmt.Errorf("failed to delete objects: %w", err)
		}

		deletedCount += len(batch)
	}

	fmt.Printf("\nSuccessfully deleted %d old backup(s) from S3.\n", deletedCount)
	return deletedCount, nil
}
