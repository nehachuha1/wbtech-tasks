package config

import (
	"os"
	"sync"
	"time"
)

// Конфиг для работы с Postgres
type PostgresConfig struct {
	PostgresUser     string
	PostgresPassword string
	PostgresAddress  string
	PostgresPort     string
	PostgresDatabase string
}

// Конфиг для работы с кафкой
type KafkaConfig struct {
	KafkaURL string
	Topic    string
}

// Конфиг для хранилища кэша
type CacheConfig struct {
	ClearInterval time.Duration
	CacheLimit    int64
	mu            sync.RWMutex
}

// В инициализация конфигов подгружаются переменные окружения. Если невозможно найти переменные окружения,
// в поля структур присваиваются "дефолтные" значения.

// Инициализация нового конфига для Postgres
func NewPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		PostgresUser:     getFromEnv("POSTGRES_USER", "admin"),
		PostgresPassword: getFromEnv("POSTGRES_PASSWORD", "nasud2198vsd2dv"),
		PostgresAddress:  getFromEnv("POSTGRES_ADDRESS", "localhost"),
		PostgresPort:     getFromEnv("POSTGRES_PORT", "5432"),
		PostgresDatabase: getFromEnv("POSTGRES_DATABASE", "maindb"),
	}
}

// Инициализация нового конфига для Kafka
func NewKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		KafkaURL: getFromEnv("KAFKA_URL", "127.0.0.1:9092"),
		Topic:    getFromEnv("KAFKA_TOPIC", "orders"),
	}
}

// Инициализация нового конфига для хранилища кэша. Очищение хранилища происходит каждые 30 минут. Лимит по количеству
// записей - 5000
func NewCacheConfig() *CacheConfig {
	return &CacheConfig{
		ClearInterval: time.Minute * time.Duration(30),
		CacheLimit:    5000,
	}
}

// Вспомогательная функция для получения переменной окружения
func getFromEnv(key string, defaultValue string) string {
	if value, isExists := os.LookupEnv(key); isExists {
		return value
	}
	return defaultValue
}
