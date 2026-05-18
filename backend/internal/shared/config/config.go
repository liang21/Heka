package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// DatabaseConfig holds PostgreSQL connection parameters.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig holds Redis connection parameters.
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// MilvusConfig holds Milvus vector database connection parameters.
type MilvusConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// JWTConfig holds JWT authentication parameters.
type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

// ServerConfig holds HTTP server parameters.
type ServerConfig struct {
	Port           int      `mapstructure:"port"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// ProviderConfig holds a single AI provider configuration.
type ProviderConfig struct {
	Name        string  `mapstructure:"name"`
	APIKey      string  `mapstructure:"api_key"`
	BaseURL     string  `mapstructure:"base_url"`
	Model       string  `mapstructure:"model"`
	Priority    int     `mapstructure:"priority"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// AIConfig holds AI provider configurations.
type AIConfig struct {
	Providers []ProviderConfig `mapstructure:"providers"`
}

// UploadConfig holds file upload parameters.
type UploadConfig struct {
	MaxSize      int64    `mapstructure:"max_size"`
	AllowedTypes []string `mapstructure:"allowed_types"`
	StoragePath  string   `mapstructure:"storage_path"`
}

// Config is the root configuration struct embedding all sub-configs.
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Milvus   MilvusConfig   `mapstructure:"milvus"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Server   ServerConfig   `mapstructure:"server"`
	AI       AIConfig       `mapstructure:"ai"`
	Upload   UploadConfig   `mapstructure:"upload"`
}

// Docker Compose legacy env var mappings
var legacyMappings = map[string]string{
	"DATABASE_HOST":      "HEKA_DB_HOST",
	"DATABASE_PORT":      "HEKA_DB_PORT",
	"DATABASE_USER":      "HEKA_DB_USER",
	"DATABASE_PASSWORD":  "HEKA_DB_PASSWORD",
	"DATABASE_NAME":      "HEKA_DB_NAME",
	"REDIS_HOST":         "HEKA_REDIS_HOST",
	"REDIS_PORT":         "HEKA_REDIS_PORT",
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetEnvPrefix("HEKA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Map legacy env vars
	for old, newKey := range legacyMappings {
		if val := v.GetString(old); val != "" {
			v.Set(newKey, val)
		}
	}

	// Set defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", "5m")
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("milvus.host", "localhost")
	v.SetDefault("milvus.port", 19530)
	v.SetDefault("jwt.access_token_ttl", "24h")
	v.SetDefault("jwt.refresh_token_ttl", "168h")
	v.SetDefault("server.port", 8080)
	v.SetDefault("upload.max_size", "52428800")
	v.SetDefault("upload.storage_path", "./uploads")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
