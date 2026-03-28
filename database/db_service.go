package database

import (
	"fmt"
	"mysql-database-backup-manager/configs"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func DumpDatabase(db configs.DBConfig, backupDir string) (string, error) {
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
