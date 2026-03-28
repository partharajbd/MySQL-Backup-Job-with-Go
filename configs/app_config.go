package configs

type AppConfig struct {
	BackupDir       string
	BackupRetention int
	S3Bucket        string
	Region          string
	S3Endpoint      string
	S3RootDir       string
}
