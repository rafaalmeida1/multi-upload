package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	JWTSecret    string
	Port         string
	UploadPath   string
	Environment  string
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	ContactEmail string
	FromEmail    string
	FromName     string
}

func Load() *Config {
	return &Config{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5433"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "postgres"),
		DBName:       getEnv("DB_NAME", "multiupload"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
		Port:         getEnv("PORT", "8082"),
		UploadPath:   getEnv("UPLOAD_PATH", "./uploads"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		SMTPHost:     getEnv("SMTP_HOST", "smtp.sendgrid.net"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		ContactEmail: getEnv("CONTACT_EMAIL", "comercialjam@zohomail.com"),
		FromEmail:    getEnv("FROM_EMAIL", "comercialjam@zohomail.com"),
		FromName:     getEnv("FROM_NAME", "JAM Locação de Guindastes"),
	}
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
