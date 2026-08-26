package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TempiexGrpcAddress string     `yaml:"tempiexGrpcAddress"`
	Host               string     `yaml:"host"`
	Port               int        `yaml:"port"`
	DbDsn              string     `yaml:"dbDsn"`
	Auth               AuthConfig `yaml:"auth"`
	Cors               CorsConfig `yaml:"cors"`
}

type AuthConfig struct {
	Enabled bool `yaml:"enabled"`
}

type CorsConfig struct {
	CookieInsecure bool `yaml:"cookieInsecure"`
}

func Load() (*Config, error) {
	path := os.Getenv("TEMPIEX_CONFIG")
	if path == "" {
		path = "config/development.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	return &cfg, nil
}
