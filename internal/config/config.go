package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration.
type Config struct {
	DB      DBConfig
	Kafka   KafkaConfig
	Service ServiceConfig
}

// DBConfig holds the database configuration.
type DBConfig struct {
	User     string
	Password string
	Name     string
	Host     string
	Port     string
}

// KafkaConfig holds the Kafka configuration.
type KafkaConfig struct {
	Broker string
	Topic  string
	Group  string
}

// ServiceConfig holds the service-specific configuration.
type ServiceConfig struct {
	CacheSize int
	HTTPPort  string
}

// LoadConfig loads configuration from environment variables and returns a Config object.
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found")
	}

	cacheSize, err := strconv.Atoi(getEnv("CACHE_SIZE", "5"))
	if err != nil {
		log.Fatalf("CACHE_SIZE: %v", err)
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
			Broker: getEnv("KAFKA_BROKER", "host.docker.internal:9092"),
			Topic:  getEnv("KAFKA_TOPIC", "orders"),
			Group:  getEnv("KAFKA_GROUP", "order-consumer-group"),
		},
		Service: ServiceConfig{
			CacheSize: cacheSize,
			HTTPPort:  getEnv("HTTP_PORT", "8081"),
		},
	}
}

// getEnv returns the value of the environment variable or the default value if not set.
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
