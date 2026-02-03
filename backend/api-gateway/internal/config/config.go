// Package config provides configuration management for API Gateway
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	LDAP     LDAPConfig     `yaml:"ldap"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string        `yaml:"host" env:"SERVER_HOST" default:"0.0.0.0"`
	Port            int           `yaml:"port" env:"SERVER_PORT" default:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT" default:"15s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT" default:"15s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SERVER_SHUTDOWN_TIMEOUT" default:"10s"`
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string `yaml:"host" env:"DB_HOST" default:"localhost"`
	Port     int    `yaml:"port" env:"DB_PORT" default:"5432"`
	User     string `yaml:"user" env:"DB_USER" default:"myops"`
	Password string `yaml:"password" env:"DB_PASSWORD" default:"myops_dev_pass"`
	Database string `yaml:"database" env:"DB_NAME" default:"myops"`
}

// DSN returns the database connection string
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Database)
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr     string `yaml:"addr" env:"REDIS_ADDR" default:"localhost:6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD" default:""`
	DB       int    `yaml:"db" env:"REDIS_DB" default:"0"`
	PoolSize int    `yaml:"pool_size" env:"REDIS_POOL_SIZE" default:"10"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	PrivateKeyPath  string        `yaml:"private_key_path" env:"JWT_PRIVATE_KEY_PATH" default:"./keys/private.pem"`
	PublicKeyPath   string        `yaml:"public_key_path" env:"JWT_PUBLIC_KEY_PATH" default:"./keys/public.pem"`
	Issuer          string        `yaml:"issuer" env:"JWT_ISSUER" default:"myops"`
	AccessDuration  time.Duration `yaml:"access_duration" env:"JWT_ACCESS_DURATION" default:"1h"`
	RefreshDuration time.Duration `yaml:"refresh_duration" env:"JWT_REFRESH_DURATION" default:"720h"`
}

// LDAPConfig holds LDAP configuration
type LDAPConfig struct {
	URL          string `yaml:"url" env:"LDAP_URL" default:""`
	BindDN       string `yaml:"bind_dn" env:"LDAP_BIND_DN" default:""`
	BindPassword string `yaml:"bind_password" env:"LDAP_BIND_PASSWORD" default:""`
	BaseDN       string `yaml:"base_dn" env:"LDAP_BASE_DN" default:""`
	UserFilter   string `yaml:"user_filter" env:"LDAP_USER_FILTER" default:"(uid=%s)"`
}

// Load loads configuration from file and environment variables
func Load(path string) (*Config, error) {
	cfg := &Config{}

	// Set defaults
	cfg.Server = ServerConfig{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
	cfg.Database = DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "myops",
		Password: "myops_dev_pass",
		Database: "myops",
	}
	cfg.Redis = RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize: 10,
	}
	cfg.JWT = JWTConfig{
		PrivateKeyPath:  "./keys/private.pem",
		PublicKeyPath:   "./keys/public.pem",
		Issuer:          "myops",
		AccessDuration:  1 * time.Hour,
		RefreshDuration: 720 * time.Hour, // 30 days
	}
	cfg.LDAP = LDAPConfig{
		URL:          "",
		BindDN:       "",
		BindPassword: "",
		BaseDN:       "dc=example,dc=com",
		UserFilter:   "(uid=%s)",
	}

	// Load from file if provided
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	// Override with environment variables
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = i
		}
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("LDAP_URL"); v != "" {
		cfg.LDAP.URL = v
	}
	if v := os.Getenv("LDAP_BIND_DN"); v != "" {
		cfg.LDAP.BindDN = v
	}
	if v := os.Getenv("LDAP_BIND_PASSWORD"); v != "" {
		cfg.LDAP.BindPassword = v
	}
	if v := os.Getenv("LDAP_BASE_DN"); v != "" {
		cfg.LDAP.BaseDN = v
	}
	if v := os.Getenv("LDAP_USER_FILTER"); v != "" {
		cfg.LDAP.UserFilter = v
	}

	return cfg, nil
}
