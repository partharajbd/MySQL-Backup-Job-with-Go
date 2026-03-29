# MySQL Database Backup Manager

A robust, concurrent MySQL database backup solution with S3-compatible storage support and automated retention policies.
Built with Go for high performance and reliability.

## Author

**Project Maintainer**: Partharaj Deb  
**Email**: mail@partharaj.me  
**GitHub**: [@partharajbd](https://github.com/partharajbd)

## Features

- ✅ **Concurrent Backups**: Utilizes worker pools to backup multiple databases simultaneously
- ✅ **S3-Compatible Storage**: Supports AWS S3 and any S3-compatible storage (MinIO, DigitalOcean Spaces, etc.)
- ✅ **Automated Retention**: Configurable retention policy to automatically delete old backups
- ✅ **Slack Notifications**: Real-time backup status notifications via Slack webhooks
- ✅ **Email Notifications**: Support for email-based notifications (coming soon)
- ✅ **Flexible Configuration**: JSON-based database configuration with environment variable support
- ✅ **Error Handling**: Comprehensive error tracking and reporting
- ✅ **Local Cleanup**: Automatic cleanup of local backup files after S3 upload
- ✅ **Summary Reports**: Detailed backup summary with success/failure counts and duration

## Prerequisites

Before installing, ensure you have the following:

- Go 1.19 or higher
- MySQL/MariaDB database(s) to backup
- `mysqldump` utility installed on your system
- AWS S3 or S3-compatible storage credentials
- (Optional) Slack webhook URL for notifications

## Installation

### 1. Clone the Repository

```bash
git clone git@github.com:partharajbd/MySQL-Backup-Job-with-Go.git
cd mysql-database-backup-manager
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment Variables

Create a `.env` file in the project root:

```bash
cp .env.example .env
```

Edit the `.env` file with your configuration:

```env
# Backup Configuration
BACKUP_DIR=./backups
RETENTION_DAYS=7

# S3 Configuration
S3_BUCKET=your-bucket-name
AWS_REGION=us-east-1
S3_ENDPOINT=https://s3.amazonaws.com  # Optional: for S3-compatible storage
S3_ROOT_DIR=mysql-backups

# AWS Credentials (set via environment or AWS credentials file)
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key

# Slack Notifications (Optional)
SLACK_ENABLED=true
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL

# Server Name (for notifications)
SERVER_NAME=production-server

# Database Configuration (JSON array)
# The value of this variable should be written in a single line, formatted as JSON here for cleaner readability
DATABASES='[
  {
    "name": "database1",
    "user": "db_user",
    "pass": "db_password",
    "host": "localhost",
    "port": "3306"
  },
  {
    "name": "database2",
    "user": "db_user",
    "pass": "db_password",
    "host": "localhost",
    "port": "3306"
  }
]'
```

### 4. Build the Application

```bash
go build -o mysql-database-backup-manager
```
_**Note** You may need to build differently for different OS. Find it yourself for the target server OS (I assume you are aware of this)._
Example for linux amd64: `GOOS=linux GOARCH=amd64 go build -o mysql-database-backup-manager`

### 5. Run the Backup

```bash
./mysql-database-backup-manager
```

## Configuration

### Database Configuration

The `DATABASES` environment variable accepts a JSON array of database configurations:

```json
[
  {
    "name": "database_name",
    "user": "username",
    "pass": "password",
    "host": "hostname",
    "port": "3306"
  }
]
```

### S3 Storage Options

**AWS S3:**
```env
S3_BUCKET=my-backup-bucket
AWS_REGION=us-east-1
S3_ENDPOINT=  # Leave empty for AWS S3
```

**MinIO or S3-Compatible:**
```env
S3_BUCKET=my-backup-bucket
AWS_REGION=us-east-1
S3_ENDPOINT=https://minio.example.com
```

### Retention Policy

Set the number of days to retain backups:

```env
RETENTION_DAYS=7
```

Backups older than this will be automatically deleted from S3.

## Usage

### Manual Backup

Run a one-time backup:

```bash
./mysql-database-backup-manager
```

### Automated Backups with Cron

Add to crontab for daily backups at 2 AM:

```bash
crontab -e
```

Add the following line:

```cron
0 2 * * * cd /path/to/mysql-database-backup-manager && ./mysql-database-backup-manager >> /var/log/mysql-backup.log 2>&1
```

### Docker Deployment [Not tested]

Create a `Dockerfile`:

```dockerfile
FROM golang:1.19-alpine AS builder

RUN apk add --no-cache git mysql-client

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o mysql-database-backup-manager

FROM alpine:latest
RUN apk add --no-cache mysql-client ca-certificates

WORKDIR /app
COPY --from=builder /app/mysql-database-backup-manager .

CMD ["./mysql-database-backup-manager"]
```

Build and run:

```bash
docker build -t mysql-backup-manager .
docker run --env-file .env mysql-backup-manager
```

## Slack Notifications

The application sends notifications for:

- **Backup Summary**: Total databases processed, success/failure counts
- **Error Alerts**: Critical errors during configuration or backup
- **Cleanup Report**: Number of old backups deleted


---

## Support

If you find this project useful, consider contributing or sharing feedback.

**Contributors are welcome.**  
If you’d like to help improve backup workflows, storage integrations, or notification support, feel free to join in.
