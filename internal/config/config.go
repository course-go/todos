package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Service struct {
	Name     string `yaml:"name,omitempty"`
	Host     string `yaml:"host,omitempty"`
	Port     string `yaml:"port,omitempty"`
	Location string `yaml:"location,omitempty"`
}

type Logging struct {
	Level string `yaml:"level,omitempty"`
}

type Database struct {
	Protocol string `yaml:"protocol"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	Options  string `yaml:"options,omitempty"`
}

type Config struct {
	Service  `yaml:"service,omitempty"`
	Logging  `yaml:"logging,omitempty"`
	Database `yaml:"database"`
}

func Parse(configPath string) (config *Config, err error) {
	var cfg Config

	file, err := os.Open(configPath)
	if err != nil {
		err = fmt.Errorf("failed opening config file: %w", err)
		return
	}

	configBytes, err := io.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("failed reading config file: %w", err)
		return
	}

	err = yaml.Unmarshal(configBytes, &cfg)
	if err != nil {
		err = fmt.Errorf("failed unmarshalling config file: %w", err)
		return
	}

	setDefaults(&cfg)
	config = &cfg

	return
}

func setDefaults(cfg *Config) {
	if cfg.Service.Name == "" {
		cfg.Service.Name = "unknown"
	}

	if cfg.Service.Port == "" {
		cfg.Service.Name = "8080"
	}

	if cfg.Location == "" {
		cfg.Location = "Local"
	}

	if cfg.Level == "" {
		cfg.Level = "info"
	}
}
