package config

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	JWT           JWTConfig
	I18n          I18nConfig
	Log           LogConfig
	Notifications NotificationsConfig
	Kafka         KafkaConfig
	CORS          CORSConfig
	Tracing       TracingConfig
}

type NotificationsConfig struct {
	ServiceURL string
}

type TracingConfig struct {
	Enabled        bool
	OTLPEndpoint   string
	ServiceName    string
	ServiceVersion string
	Environment    string
}

type KafkaConfig struct {
	Enabled bool
	Brokers []string
	GroupID string
}

type ServerConfig struct {
	Port    string
	Host    string
	GinMode string
}

type DatabaseConfig struct {
	URL         string
	AutoMigrate bool
}

type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type I18nConfig struct {
	DefaultLocale string
}

type LogConfig struct {
	Level string
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	kafkaBrokers := getEnv("KAFKA_BROKERS", "")
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "")

	ginMode := getEnv("GIN_MODE", "debug")
	jwtSecret := getEnv("JWT_SECRET", "")
	if ginMode == "release" && jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required in production mode")
	}
	if jwtSecret == "" {
		log.Println("WARNING: Using default JWT_SECRET. This is insecure and should only be used in development.")
		jwtSecret = "dev-secret-do-not-use-in-production"
	}

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			Host:    getEnv("SERVER_HOST", "0.0.0.0"),
			GinMode: ginMode,
		},
		Database: DatabaseConfig{
			URL:         getEnvRequired("DATABASE_URL"),
			AutoMigrate: getEnv("AUTO_MIGRATE", "true") == "true",
		},
		JWT: JWTConfig{
			Secret:               jwtSecret,
			AccessTokenDuration:  parseDuration(getEnv("JWT_ACCESS_TOKEN_DURATION", "15m")),
			RefreshTokenDuration: parseDuration(getEnv("JWT_REFRESH_TOKEN_DURATION", "168h")),
		},
		I18n: I18nConfig{
			DefaultLocale: getEnv("DEFAULT_LOCALE", "en-US"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Notifications: NotificationsConfig{
			ServiceURL: getEnv("NOTIFICATIONS_SERVICE_URL", "http://localhost:8081"),
		},
		Kafka: KafkaConfig{
			Enabled: getEnv("KAFKA_ENABLED", "true") == "true",
			Brokers: splitNonEmpty(kafkaBrokers, ","),
			GroupID: getEnv("KAFKA_GROUP_ID", "identity-service"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   splitNonEmpty(corsOrigins, ","),
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization", "Accept-Language", "X-Service-Name"},
			AllowCredentials: getEnv("CORS_ALLOW_CREDENTIALS", "true") == "true",
			MaxAge:           86400,
		},
		Tracing: TracingConfig{
			Enabled:        getEnv("TRACING_ENABLED", "false") == "true",
			OTLPEndpoint:   getEnv("OTLP_ENDPOINT", "localhost:4317"),
			ServiceName:    "identity",
			ServiceVersion: getEnv("SERVICE_VERSION", "1.0.0"),
			Environment:    getEnv("ENVIRONMENT", "development"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

func splitNonEmpty(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}
