package configs

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                string
	MongoURI           string
	DBName             string
	JWTSecret          []byte
	CORSOrigins        []string
	Environment        string
	RateLimitRequests  int
	RateLimitWindow    time.Duration
}

func Load() *Config {
	config := &Config{
		Port:        getEnv("PORT", "8080"),
		MongoURI:    getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DBName:      getEnv("DB_NAME", "monmetrics"),
		JWTSecret:   []byte(getEnv("JWT_SECRET", "change-this-super-secret-key")),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	// Parse CORS origins
	corsOrigins := getEnv("CORS_ORIGINS", "http://localhost:3000")
	config.CORSOrigins = strings.Split(corsOrigins, ",")

	// Parse rate limiting config
	rateLimitRequests, err := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "100"))
	if err != nil {
		rateLimitRequests = 100
	}
	config.RateLimitRequests = rateLimitRequests

	rateLimitWindow, err := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "60s"))
	if err != nil {
		rateLimitWindow = 60 * time.Second
	}
	config.RateLimitWindow = rateLimitWindow

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}