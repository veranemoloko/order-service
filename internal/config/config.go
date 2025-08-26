package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	DB      DBConfig
	Kafka   KafkaConfig
	Service ServiceConfig
}

// DBConfig holds the database configuration
type DBConfig struct {
	User     string
	Password string
	Name     string
	Host     string
	Port     string
}

// KafkaConfig holds the Kafka configuration
type KafkaConfig struct {
	Broker   string
	Topic    string
	TopicDLQ string
	Group    string
}

// ServiceConfig holds the service-specific configuration
type ServiceConfig struct {
	CacheSize int
	HTTPPort  string
	LogLevel  slog.Level
	LogFormat string
}

// LoadConfig loads configuration from environment variables and returns a config
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, using environment variables")
	}

	cacheSize, err := strconv.Atoi(getEnv("CACHE_SIZE", "5"))
	if err != nil {
		slog.Error("Invalid CACHE_SIZE", "err", err)
		os.Exit(1)
	}

	return &Config{
		DB: DBConfig{
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "orders"),
			Host:     getEnv("DB_HOST", "host.docker.internal"),
			Port:     getEnv("DB_PORT", "5432"),
		},
		Kafka: KafkaConfig{
			Broker:   getEnv("KAFKA_BROKER", "host.docker.internal:9092"),
			Topic:    getEnv("KAFKA_TOPIC", "orders"),
			TopicDLQ: getEnv("KAFKA_TOPIC_DLQ", "orders_dlq"),
			Group:    getEnv("KAFKA_GROUP", "order-consumer-group"),
		},
		Service: ServiceConfig{
			CacheSize: cacheSize,
			HTTPPort:  getEnv("HTTP_PORT", "8081"),
			LogLevel:  getLogLevel("LOG_LEVEL"),
			LogFormat: getEnv("LOG_FORMAT", "json"),
		},
	}
}

// getEnv returns the value of the environment variable or the default value if not set
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getLogLevel reads LOG_LEVEL environment variable and returns slog.Level
func getLogLevel(envVar string) slog.Level {
	levelStr := os.Getenv(envVar)
	switch levelStr {
	case "DEBUG", "debug":
		return slog.LevelDebug
	case "INFO", "info":
		return slog.LevelInfo
	case "WARN", "warn":
		return slog.LevelWarn
	case "ERROR", "error":
		return slog.LevelError
	default:
		slog.Warn("Unknown LOG_LEVEL, defaulting to INFO", "value", levelStr)
		return slog.LevelInfo
	}
}
