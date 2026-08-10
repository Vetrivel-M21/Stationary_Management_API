package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                     string
	Env                      string
	DBHost                   string
	DBPort                   string
	DBUser                   string
	DBPassword               string
	DBName                   string
	JWTSecret                string
	JWTExpirationHours       int
	JWTRefreshExpirationHour int
	SMTPHost                 string
	SMTPPort                 string
	SMTPUser                 string
	SMTPPassword             string
	SMTPFrom                 string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from system environment")
	}

	return &Config{
		Port:                     getEnv("PORT", "8080"),
		Env:                      getEnv("ENV", "development"),
		DBHost:                   getEnvMulti([]string{"DB_HOST", "MYSQLHOST"}, "127.0.0.1"),
		DBPort:                   getEnvMulti([]string{"DB_PORT", "MYSQLPORT"}, "3306"),
		DBUser:                   getEnvMulti([]string{"DB_USER", "MYSQLUSER"}, "root"),
		DBPassword:               getEnvMulti([]string{"DB_PASSWORD", "MYSQLPASSWORD"}, ""),
		DBName:                   getEnvMulti([]string{"DB_NAME", "MYSQLDATABASE"}, "stationery_db"),
		JWTSecret:                getEnv("JWT_SECRET", "super-secret-jwt-key-stationery-management-2026"),
		JWTExpirationHours:       24,
		JWTRefreshExpirationHour: 168,
		SMTPHost:                 getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:                 getEnv("SMTP_PORT", "587"),
		SMTPUser:                 getEnv("SMTP_USER", "notifications@stationeryapp.com"),
		SMTPPassword:             getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:                 getEnv("SMTP_FROM", "Stationery Management <notifications@stationeryapp.com>"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func getEnvMulti(keys []string, fallback string) string {
	for _, key := range keys {
		if value, exists := os.LookupEnv(key); exists && value != "" {
			return value
		}
	}
	return fallback
}
