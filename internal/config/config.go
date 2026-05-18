package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const defaultJWTSecret = "your-secret-key-change-in-production"

type Config struct {
	AppEnv       string // development | production
	Port         string
	JWTSecret    string
	CookieDomain string
	CookieSecure bool
	DatabasePath string
	// MySQL configuration
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	// CORS configuration
	CORSAllowedOrigins   string
	CORSAllowedMethods   string
	CORSAllowedHeaders   string
	CORSExposedHeaders   string
	CORSAllowCredentials bool
	CORSMaxAge           int
	// Notifications configuration
	NotificationsEnabled      bool
	NotificationCheckInterval string
	// Email SMTP configuration
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	// Telegram Bot configuration
	TelegramBotToken string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	corsMaxAge := 3600
	if maxAgeStr := getEnv("CORS_MAX_AGE", ""); maxAgeStr != "" {
		if parsed, err := parseInt(maxAgeStr); err == nil {
			corsMaxAge = parsed
		}
	}

	corsAllowCredentials := true
	if allowCredsStr := getEnv("CORS_ALLOW_CREDENTIALS", ""); allowCredsStr != "" {
		corsAllowCredentials = allowCredsStr == "true" || allowCredsStr == "1"
	}

	notificationsEnabled := true
	if enabledStr := getEnv("NOTIFICATIONS_ENABLED", ""); enabledStr != "" {
		notificationsEnabled = enabledStr == "true" || enabledStr == "1"
	}

	appEnv := getEnv("APP_ENV", "development")
	cookieSecure := appEnv == "production"
	if secureStr := getEnv("COOKIE_SECURE", ""); secureStr != "" {
		cookieSecure = secureStr == "true" || secureStr == "1"
	}

	cfg := &Config{
		AppEnv:                    appEnv,
		Port:                      getEnv("PORT", "8080"),
		JWTSecret:                 getEnv("JWT_SECRET", defaultJWTSecret),
		CookieDomain:              getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:              cookieSecure,
		DatabasePath:              getEnv("DATABASE_PATH", "todo.db"),
		DatabaseHost:              getEnv("DATABASE_HOST", ""),
		DatabasePort:              getEnv("DATABASE_PORT", "3306"),
		DatabaseUser:              getEnv("DATABASE_USER", ""),
		DatabasePassword:          getEnv("DATABASE_PASSWORD", ""),
		DatabaseName:              getEnv("DATABASE_NAME", ""),
		CORSAllowedOrigins:        getEnv("CORS_ALLOWED_ORIGINS", "*"),
		CORSAllowedMethods:        getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS,PATCH"),
		CORSAllowedHeaders:        getEnv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization,Accept,Origin"),
		CORSExposedHeaders:        getEnv("CORS_EXPOSED_HEADERS", ""),
		CORSAllowCredentials:      corsAllowCredentials,
		CORSMaxAge:                corsMaxAge,
		NotificationsEnabled:      notificationsEnabled,
		NotificationCheckInterval: getEnv("NOTIFICATION_CHECK_INTERVAL", "0 * * * *"),
		SMTPHost:                  getEnv("SMTP_HOST", ""),
		SMTPPort:                  getEnv("SMTP_PORT", "587"),
		SMTPUser:                  getEnv("SMTP_USER", ""),
		SMTPPassword:              getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:                  getEnv("SMTP_FROM", ""),
		TelegramBotToken:          getEnv("TELEGRAM_BOT_TOKEN", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	logConfigStatus(cfg)
	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func (c *Config) Validate() error {
	if !c.IsProduction() {
		return nil
	}
	if c.JWTSecret == "" || c.JWTSecret == defaultJWTSecret {
		return errors.New("JWT_SECRET must be set to a strong unique value in production")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters in production (got %d)", len(c.JWTSecret))
	}
	origins := strings.TrimSpace(c.CORSAllowedOrigins)
	if origins == "" || origins == "*" {
		return errors.New("CORS_ALLOWED_ORIGINS must list explicit origins in production (wildcard not allowed)")
	}
	return nil
}

func (c *Config) UseMySQL() bool {
	return c.DatabaseHost != "" && c.DatabaseUser != "" && c.DatabaseName != ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func logConfigStatus(cfg *Config) {
	log.Println("=== Configuration Status ===")
	log.Printf("APP_ENV: %s", cfg.AppEnv)
	log.Printf("Port: %s", cfg.Port)
	log.Printf("CORS Allowed Origins: %s", cfg.CORSAllowedOrigins)
	log.Printf("CORS Allow Credentials: %v", cfg.CORSAllowCredentials)
	log.Printf("Cookie Secure: %v", cfg.CookieSecure)
	log.Printf("Notifications Enabled: %v", cfg.NotificationsEnabled)
	log.Printf("SMTP Host: %s", maskIfEmpty(cfg.SMTPHost))
	log.Printf("Telegram Bot Token: %s", maskIfEmpty(cfg.TelegramBotToken))
	log.Println("===========================")
}

func maskIfEmpty(s string) string {
	if s == "" {
		return "[NOT CONFIGURED]"
	}
	return "[CONFIGURED]"
}
