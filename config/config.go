package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App       `yaml:"app"`
		HTTP      `yaml:"http"`
		Log       `yaml:"logger"`
		PG        `yaml:"postgres"`
		Converter `yaml:"converter"`
		Redis     `yaml:"redis"`
	}

	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	Log struct {
		Level string `env-required:"true" yaml:"log_level" env:"LOG_LEVEL"`
	}

	PG struct {
		PoolMax int    `env-required:"true" yaml:"pool_max" env:"PG_POOL_MAX"`
		URL     string `env-required:"true"                 env:"PG_URL"`
	}

	Converter struct {
		URL    string `env-required:"true" yaml:"url" env:"CONVERTER_URL"`
		ApiKey string `env-required:"true"            env:"CONVERTER_API_KEY"`
	}

	Redis struct {
		Addr     string `env-required:"true" yaml:"addr" env:"REDIS_ADDRESS"`
		Password string `env-required:"true"             env:"REDIS_PASSWORD"`
		DB       int    `env-required:"true" yaml:"db"   env:"REDIS_DB"`
	}
)

func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yaml", cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	err = cleanenv.UpdateEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading env: %w", err)
	}

	return cfg, nil
}
