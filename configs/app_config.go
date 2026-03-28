package configs

type AppConfig struct {
	Databases  []DBConfig
	BackupDir  string
	S3Bucket   string
	Region     string
	S3Endpoint string
}
