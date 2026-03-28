package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"mysql-database-backup-manager/configs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func loadConfig() (*configs.AppConfig, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	var dbConfigs []configs.DBConfig
	err = json.Unmarshal([]byte(os.Getenv("DATABASES")), &dbConfigs)
	if err != nil {
		return nil, err
	}

	return &configs.AppConfig{
		Databases:  dbConfigs,
		BackupDir:  os.Getenv("BACKUP_DIR"),
		S3Bucket:   os.Getenv("S3_BUCKET"),
		Region:     os.Getenv("AWS_REGION"),
		S3Endpoint: os.Getenv("S3_ENDPOINT"),
	}, nil
}

func dumpDatabase(db configs.DBConfig, backupDir string) (string, error) {
	fmt.Println("Dumping database: ", db, backupDir)
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_%s.sql", db.Name, timestamp)
	path := filepath.Join(backupDir, fileName)

	cmd := exec.Command(
		"mysqldump",
		"-h", db.Host,
		"-P", db.Port,
		"-u", db.User,
		fmt.Sprintf("-p%s", db.Pass),
		db.Name,
	)

	outFile, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return fileName, nil
}

func uploadToS3(filePath string, appConfig *configs.AppConfig, key string) error {
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

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(appConfig.S3Bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	fmt.Printf("Successfully uploaded %s to s3://%s/%s\n", filePath, appConfig.S3Bucket, key)
	return nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		panic(err)
	}

	os.MkdirAll(config.BackupDir, os.ModePerm)

	for _, db := range config.Databases {
		// Dump database
		fileName, err := dumpDatabase(db, config.BackupDir)
		if err != nil {
			fmt.Printf("Error dumping database %s: %v\n", db.Name, err)
			continue
		}
		fmt.Println("Dump completed: " + db.Name)

		// Upload to S3
		filePath := filepath.Join(config.BackupDir, fileName)
		err = uploadToS3(filePath, config, fileName)
		if err != nil {
			fmt.Printf("Error uploading %s: %v\n", fileName, err)
		}
	}
}
