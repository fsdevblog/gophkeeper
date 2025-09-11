package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPServerAddr     string `env:"HTTP_SERVER_ADDR" envDefault:":8080"`
	DatabaseDSN        string `env:"DATABASE_DSN"`
	JWTSecret          string `env:"JWT_SECRET"`
	JWTExpireInSeconds int    `env:"JWT_EXPIRE" envDefault:"3600"`
}

func MustLoadConfig() *Config {
	c, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	return c
}

func LoadConfig() (*Config, error) {
	var envConf, flagsConf Config
	if errEnvParse := env.Parse(&envConf); errEnvParse != nil {
		return nil, fmt.Errorf("parse env variables: %w", errEnvParse)
	}
	loadsFlags(&flagsConf)
	return mergeConfigs(&flagsConf, &envConf), nil
}

func mergeConfigs(fl, e *Config) *Config {
	return &Config{
		HTTPServerAddr:     firstNonEmpty(fl.HTTPServerAddr, e.HTTPServerAddr),
		DatabaseDSN:        firstNonEmpty(fl.DatabaseDSN, e.DatabaseDSN),
		JWTSecret:          firstNonEmpty(fl.JWTSecret, e.JWTSecret),
		JWTExpireInSeconds: firstNonEmpty(fl.JWTExpireInSeconds, e.JWTExpireInSeconds),
	}
}

func loadsFlags(flagsConfig *Config) {
	flag.StringVar(&flagsConfig.DatabaseDSN, "d", "", "database dsn")
	flag.StringVar(&flagsConfig.JWTSecret, "s", "", "jwt secret")
	flag.StringVar(&flagsConfig.HTTPServerAddr, "a", ":8080", "server address")
	flag.Parse()
}

// firstNonEmpty возвращает первое непустое значение из списка или значение типа T по умолчанию.
func firstNonEmpty[T comparable](values ...T) T {
	var zero T
	for _, v := range values {
		if v != zero {
			return v
		}
	}
	return zero
}
