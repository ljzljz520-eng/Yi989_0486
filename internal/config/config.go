package config

import (
	"os"
	"strconv"
)

type Config struct {
	DataPath string
	Address  string
	LogLevel string
	Port     int
}

func Default() Config {
	return Config{DataPath: "./data/couponbatch.db", Address: "127.0.0.1", LogLevel: "info", Port: 8080}
}

func Load() Config {
	cfg := Default()
	if value := os.Getenv("COUPONBATCH_DATA"); value != "" {
		cfg.DataPath = value
	}
	if value := os.Getenv("COUPONBATCH_ADDR"); value != "" {
		cfg.Address = value
	}
	if value := os.Getenv("COUPONBATCH_LOG"); value != "" {
		cfg.LogLevel = value
	}
	if value := os.Getenv("COUPONBATCH_PORT"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			cfg.Port = parsed
		}
	}
	return cfg
}

func (c Config) ListenAddress() string { return c.Address + ":" + strconv.Itoa(c.Port) }
