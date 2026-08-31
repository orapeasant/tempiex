// Package config loads worker's YAML configuration (Tempiex connection
// details, polled task queues, logging), mirroring the shape of pi's own
// config but deliberately narrower: no API/session/metrics sections, since
// those are AI-agent-harness concerns that live in pi, not in the generic
// worker binary.
package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Tempiex TempiexConfig  `yaml:"tempiex"`
	Workers []WorkerConfig `yaml:"workers"`
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
				TaskQueue:                  "worker-default",
				MaxConcurrentActivities:    10,
				MaxConcurrentWorkflowTasks: 5,
			},
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
	if v := os.Getenv("WORKER_TEMPIEX_ADDRESS"); v != "" {
		cfg.Tempiex.Address = v
	}
	if v := os.Getenv("WORKER_TEMPIEX_NAMESPACE"); v != "" {
		cfg.Tempiex.Namespace = v
	}
	if v := os.Getenv("WORKER_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("WORKER_TASK_QUEUE"); v != "" && len(cfg.Workers) > 0 {
		cfg.Workers[0].TaskQueue = v
	}
}
