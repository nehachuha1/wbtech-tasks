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
}

type CacheConfig struct {
	ClearInterval time.Duration
	CacheLimit    int64
	mu            sync.RWMutex
}

func NewPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		PostgresUser:     getFromEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getFromEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresAddress:  getFromEnv("POSTGRES_ADDRESS", "localhost"),
		PostgresPort:     getFromEnv("POSTGRES_PORT", "5432"),
		PostgresDatabase: getFromEnv("POSTGRES_DATABASE", "main"),
	}
}

func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		KafkaURL: getFromEnv("KAFKA_URL", "localhost:9091"),
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