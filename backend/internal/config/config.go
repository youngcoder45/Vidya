// Package config loads and validates all runtime configuration from the
// environment. All values are prefixed APP_ and can optionally be loaded
// from a .env file (dev convenience).
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config is the root configuration struct. Construct with Load().
type Config struct {
	AppEnv   string // local | staging | production
	Port     string
	LogLevel string

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret            string
	JWTAccessTTL         time.Duration
	JWTRefreshTTL        time.Duration
	EncryptionKey        string // 32-byte AES key for PII at rest (hex or base64)
	RazorpayKeyID        string
	RazorpayKeySecret    string
	RazorpayWebhookSecret string

	CORSOrigins    []string
	SeedOnStart    bool
	RateLimitPerMin int
	AuthRateLimitPerMin int
}

// Load reads the environment (optionally a .env file) and validates it.
func Load() (*Config, error) {
	_ = godotenv.Load() // ignore error: .env is optional

	cfg := &Config{
		AppEnv:   get("APP_ENV", "local"),
		Port:     get("APP_PORT", "8080"),
		LogLevel: get("APP_LOG_LEVEL", "info"),

		DBHost:     get("APP_DB_HOST", "localhost"),
		DBPort:     getInt("APP_DB_PORT", 5432),
		DBUser:     get("APP_DB_USER", "schoolos"),
		DBPassword: get("APP_DB_PASSWORD", "schoolos"),
		DBName:     get("APP_DB_NAME", "schoolos"),
		DBSSLMode:  get("APP_DB_SSLMODE", "disable"),

		RedisAddr:     get("APP_REDIS_ADDR", "localhost:6379"),
		RedisPassword: get("APP_REDIS_PASSWORD", ""),
		RedisDB:       getInt("APP_REDIS_DB", 0),

		JWTSecret:            get("APP_JWT_SECRET", "dev-only-change-me"),
		JWTAccessTTL:         getDuration("APP_JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:        getDuration("APP_JWT_REFRESH_TTL", 30*24*time.Hour),
		EncryptionKey:        get("APP_ENC_KEY", ""),
		RazorpayKeyID:        get("APP_RAZORPAY_KEY_ID", ""),
		RazorpayKeySecret:    get("APP_RAZORPAY_KEY_SECRET", ""),
		RazorpayWebhookSecret: get("APP_RAZORPAY_WEBHOOK_SECRET", ""),

		CORSOrigins:     split(get("APP_CORS_ORIGINS", "*")),
		SeedOnStart:     getBool("APP_SEED_ON_START", false),
		RateLimitPerMin: getInt("APP_RATE_LIMIT_PER_MIN", 120),
		AuthRateLimitPerMin: getInt("APP_AUTH_RATE_LIMIT_PER_MIN", 5),
	}

	if cfg.AppEnv == "production" && len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("config: APP_JWT_SECRET must be >= 32 chars in production")
	}
	if cfg.AppEnv == "production" && cfg.EncryptionKey == "" {
		return nil, fmt.Errorf("config: APP_ENC_KEY required in production (32-byte AES key)")
	}
	if cfg.AppEnv == "production" && cfg.RazorpayWebhookSecret == "" {
		return nil, fmt.Errorf("config: APP_RAZORPAY_WEBHOOK_SECRET required in production")
	}
	if cfg.AppEnv == "production" && (cfg.RazorpayKeyID == "" || cfg.RazorpayKeySecret == "") {
		return nil, fmt.Errorf("config: APP_RAZORPAY_KEY_ID and APP_RAZORPAY_KEY_SECRET required in production")
	}
	if cfg.RateLimitPerMin <= 0 || cfg.AuthRateLimitPerMin <= 0 {
		return nil, fmt.Errorf("config: rate limits must be > 0")
	}
	if cfg.AppEnv == "production" && len(cfg.CORSOrigins) == 1 && cfg.CORSOrigins[0] == "*" {
		return nil, fmt.Errorf("config: APP_CORS_ORIGINS must be an explicit allow-list in production")
	}
	return cfg, nil
}

func get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	v := get(key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getBool(key string, def bool) bool {
	v := get(key, "")
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getDuration(key string, def time.Duration) time.Duration {
	v := get(key, "")
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func split(s string) []string {
	if s == "" || s == "*" {
		return []string{"*"}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
