package config

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
	return DatabaseConfig{
		Username: "postgres",
		Password: "NNA2s*123",
		Host:     "localhost",
		Port:     "15432",
		Database: "novaDB",
		SSLMode:  "disable",
	}
}

// LoadRedisConfig загружает настройки Redis из конфигурации
func LoadRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:             "localhost:16379",
		DB:               0,
		DBTelegramEvents: 1,
	}
}
