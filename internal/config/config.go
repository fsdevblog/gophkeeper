package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPServerAddr string `env:"HTTP_SERVER_ADDR" envDefault:":8080"`
	DatabaseDSN    string `env:"DATABASE_DSN"`
	JWTSecret      string `env:"JWT_SECRET"`
}

func MustLoadConfig() *Config {
	c, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	return c
}

func LoadConfig() (*Config, error) {
	var conf Config
	if errEnvParse := env.Parse(&conf); errEnvParse != nil {
		return nil, fmt.Errorf("parse env variables: %w", errEnvParse)
	}
	return &conf, nil
}
