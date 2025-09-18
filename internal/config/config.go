package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Database      DatabaseConfig
	Redis         RedisConfig
	Elasticsearch ElasticsearchConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host string
	Port int
}

type ElasticsearchConfig struct {
	URL string
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Database configuration
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.User = getEnv("DB_USER", "blog_user")
	cfg.Database.Password = getEnv("DB_PASSWORD", "blog_password")
	cfg.Database.DBName = getEnv("DB_NAME", "blog_db")

	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}
	cfg.Database.Port = port

	// Redis configuration
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %v", err)
	}
	cfg.Redis.Port = redisPort

	// Elasticsearch configuration
	cfg.Elasticsearch.URL = getEnv("ELASTICSEARCH_URL", "http://localhost:9200")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
