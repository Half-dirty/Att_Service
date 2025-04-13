package config

import (
	"os"
)

var (
	// Конфигурация сервера
	ServerPort = getEnv("SERVER_PORT", ":3000")

	// Конфигурация базы данных
	DBHost     = getEnv("DB_HOST", "localhost")
	DBUser     = getEnv("DB_USER", "postgres")
	DBPassword = getEnv("DB_PASSWORD", "0Oijhbgvfcrd4e.")
	DBName     = getEnv("DB_NAME", "att_service_db")
	DBPort     = getEnv("DB_PORT", "5432")
	DBSSLMode  = getEnv("DB_SSLMODE", "disable")
	DBTimeZone = getEnv("DB_TIMEZONE", "UTC")

	// Секреты для JWT и администратора
	JWTSecret   = getEnv("JWT_SECRET_KEY", "default_secret")
	AdminSecret = getEnv("ADMIN_SECRET", "default_admin_secret")

	// Конфигурация SMTP для отправки email
	SMTPHost     = getEnv("SMTP_HOST", "smtp.example.com")
	SMTPPort     = getEnv("SMTP_PORT", "587")
	SMTPUser     = getEnv("SMTP_USER", "user@example.com")
	SMTPPassword = getEnv("SMTP_PASSWORD", "your_password")
	FromEmail    = getEnv("FROM_EMAIL", "noreply@example.com")
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
