package config

import (
	"os"
	"time"
)

type Config struct {
	Port             string
	DatabaseURL      string
	RedisURL         string
	TelegramBotToken string
	BotMode          string // "polling" или "webhook"
	WebhookURL       string
	WebAppURL        string
	JWTSecret        string
	JWTTTL           time.Duration

	// Gamification
	XPConfirmTranslation int
	XPEditTranslation    int
	XPCatchHoneypot      int
	XPDailyCompletion    int

	// Consensus
	ConsensusMinVotes int
}

func LoadConfig() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:secret@localhost:5432/bashkir_school?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379/0"),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		BotMode:          getEnv("BOT_MODE", "polling"),
		WebhookURL:       getEnv("WEBHOOK_URL", "https://example.com/api/v1/telegram/webhook"),
		WebAppURL:        getEnv("WEBAPP_URL", "https://example.com/app"),
		JWTSecret:        getEnv("JWT_SECRET", "bashkir-secret-key-hackathon-2026"),
		JWTTTL:           7 * 24 * time.Hour,

		XPConfirmTranslation: 10,
		XPEditTranslation:    25,
		XPCatchHoneypot:      30,
		XPDailyCompletion:    50,

		ConsensusMinVotes: 3,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
