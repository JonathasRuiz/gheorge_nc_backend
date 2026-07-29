package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	ContactForwardTo string

	OpenAIAPIKey string
	OpenAIModel  string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port: getEnv("PORT", "8080"),

		SMTPHost:         getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:         getEnv("SMTP_PORT", "587"),
		SMTPUser:         getEnv("SMTP_USER", ""),
		SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		ContactForwardTo: getEnv("CONTACT_FORWARD_TO", ""),

		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:  getEnv("OPENAI_MODEL", "gpt-4o"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
