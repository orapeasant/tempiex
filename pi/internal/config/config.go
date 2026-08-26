package config

import (
	"os"
	"strconv"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Tempiex TempiexConfig  `yaml:"tempiex"`
	Workers []WorkerConfig `yaml:"workers"`
	API     APIConfig      `yaml:"api"`
	Session SessionConfig  `yaml:"session"`
	Metrics MetricsConfig  `yaml:"metrics"`
	Log     LogConfig      `yaml:"log"`
}

type TempiexConfig struct {
	Address   string    `yaml:"address"`
	Namespace string    `yaml:"namespace"`
	TLS       TLSConfig `yaml:"tls"`
	APIKey    string    `yaml:"apiKey"`
}

type TLSConfig struct {
	Enabled bool `yaml:"enabled"`
}

type WorkerConfig struct {
	TaskQueue                  string `yaml:"taskQueue"`
	MaxConcurrentActivities    int    `yaml:"maxConcurrentActivities"`
	MaxConcurrentWorkflowTasks int    `yaml:"maxConcurrentWorkflowTasks"`
}

type APIConfig struct {
	Host        string   `yaml:"host"`
	Port        int      `yaml:"port"`
	CORSOrigins []string `yaml:"corsOrigins"`
}

type SessionConfig struct {
	StorePath            string `yaml:"storePath"`
	MaxCompletedSessions int    `yaml:"maxCompletedSessions"`
}

type MetricsConfig struct {
	OTel OTelConfig `yaml:"otel"`
}

type OTelConfig struct {
	Endpoint string `yaml:"endpoint"`
	Insecure bool   `yaml:"insecure"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Default() *Config {
	return &Config{
		Tempiex: TempiexConfig{
			Address:   "localhost:8133",
			Namespace: "default",
		},
		Workers: []WorkerConfig{
			{
				TaskQueue:                  "pi-default",
				MaxConcurrentActivities:    10,
				MaxConcurrentWorkflowTasks: 5,
			},
		},
		API: APIConfig{
			Host: "0.0.0.0",
			Port: 8090,
		},
		Session: SessionConfig{
			StorePath:            "./pi-sessions.db",
			MaxCompletedSessions: 1000,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("PI_TEMPIEX_ADDRESS"); v != "" {
		cfg.Tempiex.Address = v
	}
	if v := os.Getenv("PI_TEMPIEX_NAMESPACE"); v != "" {
		cfg.Tempiex.Namespace = v
	}
	if v := os.Getenv("PI_API_HOST"); v != "" {
		cfg.API.Host = v
	}
	if v := os.Getenv("PI_API_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.API.Port = p
		}
	}
	if v := os.Getenv("PI_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
}
