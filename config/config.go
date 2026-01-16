package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	Env      string
	Database DatabaseConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	JWT      JWTConfig
	Google   GoogleOAuthConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Queue    string
}

type JWTConfig struct {
	Secret      string
	ExpiryHours time.Duration
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	env := getEnv("ENV", "development")

	dbConfig := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "gamemanager"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	redisConfig := RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       redisDB,
	}

	rabbitMQConfig := RabbitMQConfig{
		Host:     getEnv("RABBITMQ_HOST", "localhost"),
		Port:     getEnv("RABBITMQ_PORT", "5672"),
		User:     getEnv("RABBITMQ_USER", "guest"),
		Password: getEnv("RABBITMQ_PASSWORD", "guest"),
		Queue:    getEnv("RABBITMQ_QUEUE", "match.queue"),
	}

	expiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	jwtConfig := JWTConfig{
		Secret:      getEnv("JWT_SECRET", "supersecretkey"),
		ExpiryHours: time.Duration(expiryHours) * time.Hour,
	}

	googleConfig := GoogleOAuthConfig{
		ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
	}

	return &Config{
		Port:     port,
		Env:      env,
		Database: dbConfig,
		Redis:    redisConfig,
		RabbitMQ: rabbitMQConfig,
		JWT:      jwtConfig,
		Google:   googleConfig,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
