package configs

type AppConfig struct {
	BackupDir       string
	S3Bucket        string
	Region          string
	S3Endpoint      string
	S3RootDir       string
	SlackWebhookURL string
	SlackEnabled    bool
	ServerName      string
}
