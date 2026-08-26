package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	GRPCPort int `yaml:"grpcPort"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbName"`
	SSLMode  string `yaml:"sslMode"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{GRPCPort: 8133},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "tempiex",
			Password: "tempiex",
			DBName:   "tempiex",
			SSLMode:  "disable",
		},
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config file: %w", err)
		}
	}

	if v := os.Getenv("TEMPIEX_GRPC_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Server.GRPCPort = p
		}
	}
	if v := os.Getenv("TEMPIEX_DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("TEMPIEX_DB_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = p
		}
	}
	if v := os.Getenv("TEMPIEX_DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("TEMPIEX_DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("TEMPIEX_DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}

	return cfg, nil
}
