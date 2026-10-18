package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	RabbitMQURL    string
	StorageDir     string // local fallback
	AllowedOrigins []string
	SessionSecret  string

	// Concurrency & Resource Limits (VPS 4GB Tuning)
	DBMaxOpenConns        int
	DBMaxIdleConns        int
	MaxConcurrentRequests int
	RateLimitPerMin       int

	// Monetization & Payments
	PLISIO_API_KEY string // STRICT: Variable name must be PLISIO_API_KEY
	AppURL         string

	// Cloudflare R2 Object Storage
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string
	R2Endpoint        string
	R2PublicDomain    string
}

func Load() *Config {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgresql://postgres:supersecretpassword30charslong123@postgres:5432/media_prod?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "redis://:supersecretredispass@redis:6379")
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	storageDir := getEnv("STORAGE_DIR", "./data/uploads")
	secret := getEnv("SESSION_SECRET", "super-secret-session-key-32-bytes-long!")

	dbMaxOpen := getEnvInt("DB_MAX_OPEN_CONNS", 20)
	dbMaxIdle := getEnvInt("DB_MAX_IDLE_CONNS", 10)
	maxConcurrent := getEnvInt("MAX_CONCURRENT_REQUESTS", 300)
	rateLimitMin := getEnvInt("RATE_LIMIT_PER_MIN", 120)

	plisioApiKey := getEnv("PLISIO_API_KEY", "")
	appURL := getEnv("APP_URL", "http://localhost")

	r2AccountID := getEnv("R2_ACCOUNT_ID", "")
	r2AccessKeyID := getEnv("R2_ACCESS_KEY_ID", "")
	r2SecretAccessKey := getEnv("R2_SECRET_ACCESS_KEY", "")
	r2Bucket := getEnv("R2_BUCKET", "lostmedia")
	r2Endpoint := getEnv("R2_ENDPOINT", "")
	r2PublicDomain := getEnv("R2_PUBLIC_DOMAIN", "")

	return &Config{
		Port:                  port,
		DatabaseURL:           dbURL,
		RedisURL:              redisURL,
		RabbitMQURL:           rabbitURL,
		StorageDir:            storageDir,
		AllowedOrigins:        []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost"},
		SessionSecret:         secret,
		DBMaxOpenConns:        dbMaxOpen,
		DBMaxIdleConns:        dbMaxIdle,
		MaxConcurrentRequests: maxConcurrent,
		RateLimitPerMin:       rateLimitMin,
		PLISIO_API_KEY:        plisioApiKey,
		AppURL:                appURL,
		R2AccountID:           r2AccountID,
		R2AccessKeyID:         r2AccessKeyID,
		R2SecretAccessKey:     r2SecretAccessKey,
		R2Bucket:              r2Bucket,
		R2Endpoint:            r2Endpoint,
		R2PublicDomain:        r2PublicDomain,
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}
