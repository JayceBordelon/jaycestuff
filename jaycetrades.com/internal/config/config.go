package config

import (
	"os"
	"strings"
)

type Config struct {
	CronScheduleOpen  string // Cron expression for market open job
	CronScheduleClose string // Cron expression for market close job
	ResendAPIKey      string
	OpenAIAPIKey      string
	EmailRecipients   []string // Comma-separated list of recipients
	EmailFrom         string
	DBPath            string // Path to SQLite database file
}

func Load() *Config {
	cronOpen := os.Getenv("CRON_SCHEDULE_OPEN")
	if cronOpen == "" {
		cronOpen = "25 9 * * 1-5" // Default: 9:25am EST (5 min before market open)
	}

	cronClose := os.Getenv("CRON_SCHEDULE_CLOSE")
	if cronClose == "" {
		cronClose = "5 16 * * 1-5" // Default: 4:05pm EST (5 min after market close)
	}

	emailFrom := os.Getenv("EMAIL_FROM")
	if emailFrom == "" {
		emailFrom = "Jayce's Trading Bot <trades@jaycetrades.com>"
	}

	var recipients []string
	if r := os.Getenv("EMAIL_RECIPIENTS"); r != "" {
		for _, email := range strings.Split(r, ",") {
			if trimmed := strings.TrimSpace(email); trimmed != "" {
				recipients = append(recipients, trimmed)
			}
		}
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "jaycetrades.db"
	}

	return &Config{
		CronScheduleOpen:  cronOpen,
		CronScheduleClose: cronClose,
		ResendAPIKey:      os.Getenv("RESEND_API_KEY"),
		OpenAIAPIKey:      os.Getenv("OPENAI_API_KEY"),
		EmailRecipients:   recipients,
		EmailFrom:         emailFrom,
		DBPath:            dbPath,
	}
}
