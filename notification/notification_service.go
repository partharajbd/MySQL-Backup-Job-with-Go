package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type SlackConfig struct {
	WebhookURL string
	Enabled    bool
}

type SlackMessage struct {
	Text        string       `json:"text,omitempty"`
	Blocks      []SlackBlock `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type SlackBlock struct {
	Type string                 `json:"type"`
	Text map[string]interface{} `json:"text,omitempty"`
}

type Attachment struct {
	Color  string  `json:"color"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type BackupSummary struct {
	Total        int
	Success      int
	Failed       int
	FailedDBs    []string
	DeletedCount int
	Duration     time.Duration
	ServerName   string
}

func SendBackupSummary(config SlackConfig, summary BackupSummary) error {
	if !config.Enabled || config.WebhookURL == "" {
		return nil
	}

	status := "✅ Success"
	color := "good"
	if summary.Failed > 0 {
		status = "⚠️ Partial Success"
		color = "warning"
	}
	if summary.Success == 0 {
		status = "❌ Failed"
		color = "danger"
	}

	failedDBsText := "None"
	if len(summary.FailedDBs) > 0 {
		failedDBsText = ""
		for _, db := range summary.FailedDBs {
			failedDBsText += fmt.Sprintf("• %s\n", db)
		}
	}

	serverInfo := "Unknown"
	hostname, err := os.Hostname()
	if err == nil {
		serverInfo = hostname
	}

	message := SlackMessage{
		Text: fmt.Sprintf("%s - MySQL Backup Report", status),
		Attachments: []Attachment{
			{
				Color: color,
				Fields: []Field{
					{
						Title: "Server",
						Value: serverInfo,
						Short: true,
					},
					{
						Title: "Total Databases",
						Value: fmt.Sprintf("%d", summary.Total),
						Short: true,
					},
					{
						Title: "Timestamp",
						Value: time.Now().Format("2006-01-02 15:04:05 MST"),
						Short: true,
					},
					{
						Title: "Successful",
						Value: fmt.Sprintf("%d", summary.Success),
						Short: true,
					},
					{
						Title: "Duration",
						Value: fmt.Sprintf("%.2fs", summary.Duration.Seconds()),
						Short: true,
					},
					{
						Title: "Failed",
						Value: fmt.Sprintf("%d", summary.Failed),
						Short: true,
					},
					{
						Title: "Old Backups Deleted",
						Value: fmt.Sprintf("%d", summary.DeletedCount),
						Short: true,
					},
					{
						Title: "Failed Databases",
						Value: failedDBsText,
						Short: true,
					},
				},
			},
		},
	}

	return sendToSlack(config.WebhookURL, message)
}

func SendErrorAlert(config SlackConfig, errorMsg string, serverName string) error {
	if !config.Enabled || config.WebhookURL == "" {
		return nil
	}

	serverInfo := "Unknown"
	if serverName != "" {
		serverInfo = serverName
	}

	message := SlackMessage{
		Text: "🚨 MySQL Backup Critical Error",
		Attachments: []Attachment{
			{
				Color: "danger",
				Fields: []Field{
					{
						Title: "Server",
						Value: serverInfo,
						Short: true,
					},
					{
						Title: "Error",
						Value: errorMsg,
						Short: false,
					},
					{
						Title: "Timestamp",
						Value: time.Now().Format("2006-01-02 15:04:05 MST"),
						Short: false,
					},
				},
			},
		},
	}

	return sendToSlack(config.WebhookURL, message)
}

func sendToSlack(webhookURL string, message SlackMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack returned non-OK status: %d", resp.StatusCode)
	}

	fmt.Println("✓ Slack notification sent successfully")
	return nil
}
