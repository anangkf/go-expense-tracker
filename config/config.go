package config

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret             string
	ExpireHours        int
	RefreshSecret      string
	RefreshExpireHours int
}

type ServerConfig struct {
	Port               string
	Mode               string
	CORSAllowedOrigins []string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found!")
	}

	expireHours, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
	refreshExpireHours, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRE_HOURS", "168")) //expired after 7 days

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "expense_tracker"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", ""),
			ExpireHours:        expireHours,
			RefreshSecret:      getEnv("JWT_REFRESH_SECRET", ""),
			RefreshExpireHours: refreshExpireHours,
		},
		Server: ServerConfig{
			Port:               getEnv("SERVER_PORT", "8080"),
			Mode:               getEnv("SERVER_MODE", "debug"),
			CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "")),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// VALIDATE REQUIRED CONFIG
func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.JWT.RefreshSecret == "" {
		return errors.New("JWT_REFRESH_SECRET is required")
	}

	return nil
}

func splitCSV(value string) []string {
	origins := []string{}
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	return origins
}
