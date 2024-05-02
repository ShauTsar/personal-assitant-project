package config

import "os"

type DatabaseConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
	SSLMode  string
}

// RedisConfig содержит настройки для клиента Redis
type RedisConfig struct {
	Addr             string
	DB               int
	DBTelegramEvents int
}

// LoadDatabaseConfig загружает настройки базы данных из конфигурации
func LoadDatabaseConfig() DatabaseConfig {
	pHost := os.Getenv("POSTGRES_HOST")
	port := "5432"
	if pHost == "" {
		pHost = "localhost"
		port = "15432"
	}
	return DatabaseConfig{
		Username: "postgres",
		Password: "NNA2s*123",
		Host:     pHost,
		Port:     port,
		Database: "novaDB",
		SSLMode:  "disable",
	}
}

// LoadRedisConfig загружает настройки Redis из конфигурации
func LoadRedisConfig() RedisConfig {
	rHost := os.Getenv("REDIS_HOST")
	if rHost == "" {
		rHost = "localhost"
	}
	return RedisConfig{
		Addr:             rHost + ":16379",
		DB:               0,
		DBTelegramEvents: 1,
	}
}
