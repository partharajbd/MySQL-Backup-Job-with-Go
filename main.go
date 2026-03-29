package main

import (
	"encoding/json"
	"fmt"
	"mysql-database-backup-manager/configs"
	"mysql-database-backup-manager/database"
	"mysql-database-backup-manager/notification"
	"mysql-database-backup-manager/storage"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Job struct {
	DBConfig  configs.DBConfig
	AppConfig *configs.AppConfig
}

type Result struct {
	DBName string
	Error  error
}

func loadAppConfig() (*configs.AppConfig, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	slackEnabled := false
	if enabled := os.Getenv("SLACK_ENABLED"); enabled == "true" {
		slackEnabled = true
	}

	return &configs.AppConfig{
		BackupDir:       os.Getenv("BACKUP_DIR"),
		S3Bucket:        os.Getenv("S3_BUCKET"),
		Region:          os.Getenv("AWS_REGION"),
		S3Endpoint:      os.Getenv("S3_ENDPOINT"),
		S3RootDir:       os.Getenv("S3_ROOT_DIR"),
		SlackWebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
		SlackEnabled:    slackEnabled,
	}, nil
}

func loadDbConfig() ([]configs.DBConfig, error) {
	var dbConfigs []configs.DBConfig
	err := json.Unmarshal([]byte(os.Getenv("DATABASES")), &dbConfigs)
	if err != nil {
		return nil, err
	}

	return dbConfigs, nil
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		fileName, err := database.DumpDatabase(job.DBConfig, job.AppConfig.BackupDir)
		if err != nil {
			results <- Result{DBName: job.DBConfig.Name, Error: fmt.Errorf("dump error: %w", err)}
			continue
		}

		filePath := filepath.Join(job.AppConfig.BackupDir, fileName)
		err = storage.UploadToS3(filePath, job.AppConfig, fileName)
		if err != nil {
			results <- Result{DBName: job.DBConfig.Name, Error: fmt.Errorf("upload error: %w", err)}
			continue
		}

		if err := os.Remove(filePath); err != nil {
			results <- Result{DBName: job.DBConfig.Name, Error: fmt.Errorf("delete error: %w", err)}
			continue
		}

		results <- Result{DBName: job.DBConfig.Name, Error: nil}
	}
}

func main() {
	startTime := time.Now()

	appConfig, err := loadAppConfig()
	if err != nil {
		panic(err)
	}

	// Create Slack config
	slackConfig := notification.SlackConfig{
		WebhookURL: appConfig.SlackWebhookURL,
		Enabled:    appConfig.SlackEnabled,
	}

	dbConfigs, err := loadDbConfig()
	if err != nil {
		// Send critical error to Slack
		notification.SendErrorAlert(slackConfig, fmt.Sprintf("Failed to load database config: %v", err), appConfig.ServerName)
		panic(err)
	}

	os.RemoveAll(appConfig.BackupDir)

	numWorkers := 3
	if len(dbConfigs) < numWorkers {
		numWorkers = len(dbConfigs)
	}

	jobs := make(chan Job, len(dbConfigs))
	results := make(chan Result, len(dbConfigs))

	var wg sync.WaitGroup
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	for _, db := range dbConfigs {
		jobs <- Job{
			DBConfig:  db,
			AppConfig: appConfig,
		}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("\n=== Backup Summary ===")
	successCount := 0
	failureCount := 0
	var failedDBs []string

	for result := range results {
		if result.Error != nil {
			fmt.Printf("> Failed to process %s: %v\n", result.DBName, result.Error)
			failureCount++
			failedDBs = append(failedDBs, result.DBName)
		} else {
			fmt.Printf("> Successfully processed %s.\n", result.DBName)
			successCount++
		}
	}

	fmt.Printf("\nTotal: %d | Success: %d | Failed: %d\n",
		len(dbConfigs), successCount, failureCount)

	// Clean up old backups from S3
	fmt.Println("\n=== Cleanup Phase ===")
	deletedCount := 0
	retentionDays := 7

	// Get retention days from env if set
	if retentionEnv := os.Getenv("RETENTION_DAYS"); retentionEnv != "" {
		if days, err := strconv.Atoi(retentionEnv); err == nil {
			retentionDays = days
		}
	}

	if count, err := storage.DeleteOldBackups(appConfig, retentionDays); err != nil {
		fmt.Printf("Warning: Failed to clean old backups: %v\n", err)
	} else {
		deletedCount = count
	}

	duration := time.Since(startTime)

	// Send summary to Slack
	summary := notification.BackupSummary{
		Total:        len(dbConfigs),
		Success:      successCount,
		Failed:       failureCount,
		FailedDBs:    failedDBs,
		DeletedCount: deletedCount,
		Duration:     duration,
	}

	if err := notification.SendBackupSummary(slackConfig, summary); err != nil {
		fmt.Printf("Warning: Failed to send Slack notification: %v\n", err)
	}

	fmt.Printf("\n✓ Backup process completed in %.2fs\n", duration.Seconds())
}
