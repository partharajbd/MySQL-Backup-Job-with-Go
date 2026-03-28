package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"mysql-database-backup-manager/configs"

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
		Databases: dbConfigs,
		BackupDir: os.Getenv("BACKUP_DIR"),
		S3Bucket:  os.Getenv("S3_BUCKET"),
		Region:    os.Getenv("AWS_REGION"),
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

func main() {
	config, err := loadConfig()
	if err != nil {
		panic(err)
	}

	os.MkdirAll(config.BackupDir, os.ModePerm)

	for _, db := range config.Databases {
		dumpDatabase(db, config.BackupDir)
	}
}
