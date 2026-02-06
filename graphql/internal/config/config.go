package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	JWT Token `yaml:"jwt"`
}

type Token struct {
	Secret string
	TTL    time.Duration
}

func MustLoad(path string) *Config {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("read config: %w", err))
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		panic(fmt.Errorf("yaml unmarshal: %w", err))
	}

	return &cfg
}
