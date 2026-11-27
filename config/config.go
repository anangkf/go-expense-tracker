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
	Email    EmailConfig
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	From         string

	OTPExpireMinutes       int // LIFETIME OF A VERIFICATION CODE
	OTPMaxAttempts         int // WRONG GUESSES BEFORE THE CODE IS INVALIDATED
	ResendCooldownSeconds  int // MINIMUM GAP BETWEEN TWO SENDS
	ResendMaxPerDay        int // MAXIMUM SENDS PER 24H WINDOW
	UnverifiedUserTTLHours int // UNVERIFIED ACCOUNTS OLDER THAN THIS ARE DELETED
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
	BaseURLFE          string // FRONTEND BASE URL USED IN USER FACING MESSAGES (OPTIONAL)
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found!")
	}

	expireHours, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
	refreshExpireHours, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRE_HOURS", "168")) //expired after 7 days

	otpExpire := getEnvInt("EMAIL_OTP_EXPIRE_MINUTES", 10)
	otpMaxAttempts := getEnvInt("EMAIL_OTP_MAX_ATTEMPTS", 5)
	resendCooldown := getEnvInt("EMAIL_RESEND_COOLDOWN_SECONDS", 60)
	resendMax := getEnvInt("EMAIL_RESEND_MAX_PER_DAY", 5)
	unverifiedTTL := getEnvInt("UNVERIFIED_USER_TTL_HOURS", 168)

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
			BaseURLFE:          strings.TrimRight(strings.TrimSpace(getEnv("BASE_URL_FE", "")), "/"),
		},
		Email: EmailConfig{
			SMTPHost:               getEnv("SMTP_HOST", ""),
			SMTPPort:               getEnv("SMTP_PORT", "587"),
			SMTPUser:               getEnv("SMTP_USER", ""),
			SMTPPassword:           getEnv("SMTP_PASSWORD", ""),
			From:                   getEnv("SMTP_FROM", ""),
			OTPExpireMinutes:       otpExpire,
			OTPMaxAttempts:         otpMaxAttempts,
			ResendCooldownSeconds:  resendCooldown,
			ResendMaxPerDay:        resendMax,
			UnverifiedUserTTLHours: unverifiedTTL,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// READ A POSITIVE INTEGER ENV VAR, FALLING BACK TO DEFAULT WHEN MISSING OR INVALID
func getEnvInt(key string, defaultValue int) int {
	if n, err := strconv.Atoi(getEnv(key, "")); err == nil && n > 0 {
		return n
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

	// SMTP IS REQUIRED IN RELEASE MODE (OTHERWISE EMAILS ARE ONLY LOGGED)
	if c.Server.Mode == "release" && (c.Email.SMTPHost == "" || c.Email.From == "") {
		return errors.New("SMTP_HOST and SMTP_FROM are required when SERVER_MODE=release")
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
