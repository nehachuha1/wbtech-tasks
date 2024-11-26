package config

import (
	"os"
	"sync"
	"time"
)

type PostgresConfig struct {
	PostgresUser     string
	PostgresPassword string
	PostgresAddress  string
	PostgresPort     string
	PostgresDatabase string
}

type KafkaConfig struct {
	KafkaURL string
	Topic    string
}

type CacheConfig struct {
	ClearInterval time.Duration
	CacheLimit    int64
	mu            sync.RWMutex
}

func NewPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		PostgresUser:     getFromEnv("POSTGRES_USER", "admin"),
		PostgresPassword: getFromEnv("POSTGRES_PASSWORD", "nasud2198vsd2dv"),
		PostgresAddress:  getFromEnv("POSTGRES_ADDRESS", "localhost"),
		PostgresPort:     getFromEnv("POSTGRES_PORT", "5432"),
		PostgresDatabase: getFromEnv("POSTGRES_DATABASE", "maindb"),
	}
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		KafkaURL: getFromEnv("KAFKA_URL", "127.0.0.1:9092"),
		Topic:    getFromEnv("KAFKA_TOPIC", "orders"),
	}
}

func NewCacheConfig() *CacheConfig {
	return &CacheConfig{
		ClearInterval: time.Hour * time.Duration(1),
		CacheLimit:    1000,
	}
}

func getFromEnv(key string, defaultValue string) string {
	if value, isExists := os.LookupEnv(key); isExists {
		return value
	}
	return defaultValue
}
