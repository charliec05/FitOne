package config

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	RedisURL    string
	Environment string
	
	// Feature flags
	MapsAPIKey   string
	S3Endpoint   string
	S3Bucket     string
	S3Region     string
	MaxUploadMB  int

	ModerationEnabled bool
	CDNBaseURL        string
	CacheTTLNearby    time.Duration
	CacheTTLGym       time.Duration
	AnalyticsSink     string
	FeatureFlags      map[string]float64
	AlertWebhookURL   string
	FeatureFlagsRaw   string

	StripeSecretKey     string
	StripeWebhookSecret string
	StripeSuccessURL    string
	StripeCancelURL     string
    StripePriceID      string
	GoogleClientID      string
	GoogleClientSecret  string
	PasswordResetSecret string
	EmailSender         string
	EnvironmentName     string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	envName := getEnv("ENV_NAME", "")

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://fitonex:password@localhost:5432/fitonex?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		Environment: getEnv("ENVIRONMENT", "development"),
		
		// Feature flags
		MapsAPIKey:  getEnv("MAPS_API_KEY", ""),
		S3Endpoint:  getEnv("S3_ENDPOINT", "http://localhost:9000"),
		S3Bucket:    getEnv("S3_BUCKET", "fitonex"),
		S3Region:    getEnv("S3_REGION", "us-east-1"),
		MaxUploadMB: getEnvInt("MAX_UPLOAD_MB", 100),

		ModerationEnabled: getEnvBool("MODERATION_ENABLED", true),
		CDNBaseURL:        getEnv("CDN_BASE_URL", ""),
		CacheTTLNearby:    getEnvDuration("CACHE_TTL_NEARBY", 60*time.Second),
		CacheTTLGym:       getEnvDuration("CACHE_TTL_GYM", 5*time.Minute),
		AnalyticsSink:     getEnv("ANALYTICS_SINK", "stdout"),
		AlertWebhookURL:   getEnv("ALERT_WEBHOOK_URL", ""),
		FeatureFlagsRaw:   getEnv("FEATURE_FLAGS_JSON", `{"video_upload":100,"map_filters":50}`),
		StripeSecretKey:   getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeSuccessURL:    getEnv("STRIPE_SUCCESS_URL", "https://localhost/success"),
		StripeCancelURL:     getEnv("STRIPE_CANCEL_URL", "https://localhost/cancel"),
		StripePriceID:       getEnv("STRIPE_PRICE_ID", ""),
		GoogleClientID:      getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
		GoogleClientSecret:  getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
		PasswordResetSecret: getEnv("PASSWORD_RESET_SECRET", "change-me"),
		EmailSender:         getEnv("EMAIL_SENDER_ADDRESS", "no-reply@fitonex.local"),
	}

	if envName != "" {
		cfg.EnvironmentName = envName
	} else {
		cfg.EnvironmentName = cfg.Environment
	}

	flags := make(map[string]float64)
	if err := json.Unmarshal([]byte(cfg.FeatureFlagsRaw), &flags); err == nil {
		cfg.FeatureFlags = flags
	} else {
		cfg.FeatureFlags = map[string]float64{}
	}

	return cfg, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt gets an environment variable as integer with a fallback value
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "1", "true", "TRUE", "True":
			return true
		case "0", "false", "FALSE", "False":
			return false
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
