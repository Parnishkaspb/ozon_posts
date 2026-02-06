package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PostgreSQL PostgreSQLConfig `yaml:"postgresql"`
	JWT        Token            `yaml:"jwt"`
	GRPC       GRPC             `yaml:"grpc"`
}

type GRPC struct {
	Port int
}

type Token struct {
	Secret string
	TTL    time.Duration
}

type PostgreSQLConfig struct {
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	DB       string `yaml:"db"   env-required:"true"`
	SSLMode  string `yaml:"sslmode"  env-default:"disable"`
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

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.PostgreSQL.User,
		c.PostgreSQL.Password,
		c.PostgreSQL.Host,
		c.PostgreSQL.Port,
		c.PostgreSQL.DB,
		c.PostgreSQL.SSLMode,
	)
}
