package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/creafly/vault"
	"github.com/joho/godotenv"
)

type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	Redis          RedisConfig
	JWT            JWTConfig
	I18n           I18nConfig
	Log            LogConfig
	Notifications  NotificationsConfig
	Kafka          KafkaConfig
	CORS           CORSConfig
	Tracing        TracingConfig
	Unleash        UnleashConfig
	RateLimit      RateLimitConfig
	AccountLockout AccountLockoutConfig
	TOTPLockout    TOTPLockoutConfig
	InternalAPI    InternalAPIConfig
	MLService      MLServiceConfig
}

type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond float64
	BurstSize         int
	TrustedProxies    []string
}

type AccountLockoutConfig struct {
	Enabled         bool
	MaxAttempts     int
	LockoutDuration time.Duration
	AttemptWindow   time.Duration
}

type TOTPLockoutConfig struct {
	Enabled         bool
	MaxAttempts     int
	LockoutDuration time.Duration
	AttemptWindow   time.Duration
}

type InternalAPIConfig struct {
	APIKey          string
	AllowedServices []string
}

type MLServiceConfig struct {
	Enabled bool
	BaseURL string
	Timeout time.Duration
}

type NotificationsConfig struct {
	ServiceURL string
}

type UnleashConfig struct {
	Enabled  bool
	URL      string
	AppName  string
	APIToken string
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

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret               string
	AccessTokenSecret    string
	RefreshTokenSecret   string
	TempTokenSecret      string
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

	secrets := vault.NewSecretLoaderFromEnv("identity")

	kafkaBrokers := getEnv("KAFKA_BROKERS", "")
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "")

	ginMode := getEnv("GIN_MODE", "debug")

	jwtSecret := secrets.GetSecret("jwt_secret", "JWT_SECRET", "")
	if ginMode == "release" && jwtSecret == "" {
		log.Fatal("JWT_SECRET is required in production mode")
	}
	if jwtSecret == "" {
		log.Println("WARNING: Using default JWT_SECRET. This is insecure and should only be used in development.")
		jwtSecret = "dev-secret-do-not-use-in-production"
	}

	accessTokenSecret := secrets.GetSecret("jwt_access_secret", "JWT_ACCESS_SECRET", "")
	if accessTokenSecret == "" {
		accessTokenSecret = jwtSecret + "-access"
	}
	refreshTokenSecret := secrets.GetSecret("jwt_refresh_secret", "JWT_REFRESH_SECRET", "")
	if refreshTokenSecret == "" {
		refreshTokenSecret = jwtSecret + "-refresh"
	}
	tempTokenSecret := secrets.GetSecret("jwt_temp_secret", "JWT_TEMP_SECRET", "")
	if tempTokenSecret == "" {
		tempTokenSecret = jwtSecret + "-temp"
	}

	databaseURL := buildDatabaseURL(secrets)
	redisPassword := secrets.GetSecret("redis_password", "REDIS_PASSWORD", "")

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			Host:    getEnv("SERVER_HOST", "0.0.0.0"),
			GinMode: ginMode,
		},
		Database: DatabaseConfig{
			URL:         databaseURL,
			AutoMigrate: getEnv("AUTO_MIGRATE", "true") == "true",
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: redisPassword,
			DB:       parseInt(getEnv("REDIS_DB", "0")),
		},
		JWT: JWTConfig{
			Secret:               jwtSecret,
			AccessTokenSecret:    accessTokenSecret,
			RefreshTokenSecret:   refreshTokenSecret,
			TempTokenSecret:      tempTokenSecret,
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
		Unleash: UnleashConfig{
			Enabled:  getEnv("UNLEASH_URL", "") != "",
			URL:      getEnv("UNLEASH_URL", ""),
			AppName:  getEnv("UNLEASH_APP_NAME", "identity"),
			APIToken: getEnv("UNLEASH_API_TOKEN", ""),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getEnv("RATE_LIMIT_ENABLED", "true") == "true",
			RequestsPerSecond: parseFloat(getEnv("RATE_LIMIT_RPS", "100")),
			BurstSize:         parseInt(getEnv("RATE_LIMIT_BURST", "200")),
			TrustedProxies:    splitNonEmpty(getEnv("TRUSTED_PROXIES", ""), ","),
		},
		AccountLockout: AccountLockoutConfig{
			Enabled:         getEnv("ACCOUNT_LOCKOUT_ENABLED", "true") == "true",
			MaxAttempts:     parseInt(getEnv("ACCOUNT_LOCKOUT_MAX_ATTEMPTS", "5")),
			LockoutDuration: parseDuration(getEnv("ACCOUNT_LOCKOUT_DURATION", "15m")),
			AttemptWindow:   parseDuration(getEnv("ACCOUNT_LOCKOUT_WINDOW", "15m")),
		},
		TOTPLockout: TOTPLockoutConfig{
			Enabled:         getEnv("TOTP_LOCKOUT_ENABLED", "true") == "true",
			MaxAttempts:     parseInt(getEnv("TOTP_LOCKOUT_MAX_ATTEMPTS", "5")),
			LockoutDuration: parseDuration(getEnv("TOTP_LOCKOUT_DURATION", "5m")),
			AttemptWindow:   parseDuration(getEnv("TOTP_LOCKOUT_WINDOW", "5m")),
		},
		InternalAPI: InternalAPIConfig{
			APIKey:          secrets.GetSecret("internal_api_key", "INTERNAL_API_KEY", ""),
			AllowedServices: splitNonEmpty(getEnv("INTERNAL_ALLOWED_SERVICES", "notifications,agent,subscriptions"), ","),
		},
		MLService: MLServiceConfig{
			Enabled: getEnv("ML_SERVICE_ENABLED", "false") == "true",
			BaseURL: getEnv("ML_SERVICE_URL", "http://localhost:8090"),
			Timeout: parseDuration(getEnv("ML_SERVICE_TIMEOUT", "5s")),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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

func parseFloat(s string) float64 {
	var v float64
	_, _ = fmt.Sscanf(s, "%f", &v)
	return v
}

func parseInt(s string) int {
	var v int
	_, _ = fmt.Sscanf(s, "%d", &v)
	return v
}

func buildDatabaseURL(secrets *vault.SecretLoader) string {
	host := getEnv("DATABASE_HOST", "localhost")
	port := getEnv("DATABASE_PORT", "5432")
	name := getEnv("DATABASE_NAME", "identity")
	user := getEnv("DATABASE_USER", "postgres")
	sslMode := getEnv("DATABASE_SSL_MODE", "disable")

	password := secrets.GetSecret("database_password", "DATABASE_PASSWORD", "")
	if password == "" {
		log.Fatal("DATABASE_PASSWORD is required (from Vault or environment)")
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user,
		url.QueryEscape(password),
		host,
		port,
		name,
		sslMode,
	)
}
