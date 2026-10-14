package config

import "os"

type Config struct {
	Port string
}

func Load() Config {
	return Config{Port: env("PORT", "8080")}
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
