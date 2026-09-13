package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	SMTPHost         string
	SMTPPort         string
	SMTPUser         string
	SMTPPassword     string
	SMTPFrom         string
	ContactForwardTo string

	OpenAIAPIKey string
	OpenAIModel  string

	DBPath string

	WhatsAppToken         string
	WhatsAppPhoneNumberID string
	WhatsAppAppSecret     string
	WhatsAppVerifyToken   string
	WhatsAppTemplateName  string
	OwnerPhoneNumber      string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port: getEnv("PORT", "80"),

		SMTPHost:         getEnv("SMTP_HOST", "email-smtp.us-east-1.amazonaws.com"),
		SMTPPort:         getEnv("SMTP_PORT", "587"),
		SMTPUser:         getEnv("SMTP_USER", ""),
		SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:         getEnv("SMTP_FROM", ""),
		ContactForwardTo: getEnv("CONTACT_FORWARD_TO", ""),

		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:  getEnv("OPENAI_MODEL", "gpt-4o"),

		DBPath: getEnv("DB_PATH", "data.db"),

		WhatsAppToken:         getEnv("WHATSAPP_TOKEN", ""),
		WhatsAppPhoneNumberID: getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppAppSecret:     getEnv("WHATSAPP_APP_SECRET", ""),
		WhatsAppVerifyToken:   getEnv("WHATSAPP_VERIFY_TOKEN", ""),
		WhatsAppTemplateName:  getEnv("WHATSAPP_TEMPLATE_NAME", ""),
		OwnerPhoneNumber:      getEnv("OWNER_PHONE_NUMBER", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
