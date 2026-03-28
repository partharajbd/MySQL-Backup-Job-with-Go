package main

import (
	"encoding/json"
	"fmt"
	"mysql-database-backup-manager/configs"
	"mysql-database-backup-manager/database"
	"mysql-database-backup-manager/storage"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func loadAppConfig() (*configs.AppConfig, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	return &configs.AppConfig{
		BackupDir:  os.Getenv("BACKUP_DIR"),
		S3Bucket:   os.Getenv("S3_BUCKET"),
		Region:     os.Getenv("AWS_REGION"),
		S3Endpoint: os.Getenv("S3_ENDPOINT"),
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

func main() {
	appConfig, err := loadAppConfig()
	if err != nil {
		panic(err)
	}

	dbConfigs, err := loadDbConfig()
	if err != nil {
		panic(err)
	}

	os.MkdirAll(appConfig.BackupDir, os.ModePerm)

	for _, db := range dbConfigs {
		// Dump database
		fileName, err := database.DumpDatabase(db, appConfig.BackupDir)
		if err != nil {
			fmt.Printf("Error dumping database %s: %v\n", db.Name, err)
			continue
		}
		fmt.Println("Dump completed: " + db.Name)

		// Upload to S3
		filePath := filepath.Join(appConfig.BackupDir, fileName)
		err = storage.UploadToS3(filePath, appConfig, fileName)
		if err != nil {
			fmt.Printf("Error uploading %s: %v\n", fileName, err)
		}
	}
}
