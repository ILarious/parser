package config

import (
	"os"
	"strconv"
	"time"

	"parser/pkg/postgres"

	"github.com/joho/godotenv"
)

type Config struct {
	Postgres   postgres.Config
	Kafka      KafkaConfig
	WorkerPool WorkerPoolConfig
	VK         VKConfig
}

type KafkaConfig struct {
	Brokers             []string
	OrderRequestTopic   string
	OrderResponseTopic  string
	OrderRequestGroupID string
}

type WorkerPoolConfig struct {
	Size int
}

type VKConfig struct {
	AccessToken string
	APIVersion  string
	BaseURL     string
	Timeout     time.Duration
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Postgres: postgres.Config{
			Host:         envOrDefault("POSTGRES_HOST", "127.0.0.1"),
			Port:         envIntOrDefault("POSTGRES_PORT", 5432),
			User:         os.Getenv("POSTGRES_USER"),
			Password:     os.Getenv("POSTGRES_PASSWORD"),
			Database:     os.Getenv("POSTGRES_DB"),
			SSLMode:      envOrDefault("POSTGRES_SSLMODE", "disable"),
			MaxOpenConns: envIntOrDefault("POSTGRES_MAX_OPEN_CONNS", 20),
			MaxIdleConns: envIntOrDefault("POSTGRES_MAX_IDLE_CONNS", 20),
		},
		Kafka: KafkaConfig{
			Brokers:             []string{envOrDefault("KAFKA_BROKER", "localhost:9092")},
			OrderRequestTopic:   envOrDefault("KAFKA_ORDER_REQUEST_TOPIC", "vk-blogger-orders"),
			OrderResponseTopic:  envOrDefault("KAFKA_ORDER_RESPONSE_TOPIC", "vk-blogger-results"),
			OrderRequestGroupID: envOrDefault("KAFKA_ORDER_REQUEST_GROUP_ID", "parser"),
		},
		WorkerPool: WorkerPoolConfig{
			Size: envIntOrDefault("WORKER_POOL_SIZE", 4),
		},
		VK: VKConfig{
			AccessToken: os.Getenv("VK_ACCESS_TOKEN"),
			APIVersion:  envOrDefault("VK_API_VERSION", "5.130"),
			BaseURL:     envOrDefault("VK_BASE_URL", "https://api.vk.com/method"),
			Timeout:     time.Duration(envIntOrDefault("VK_TIMEOUT_SECONDS", 10)) * time.Second,
		},
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
