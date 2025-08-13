package config

import (
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

// Config holds the application configuration, loaded from environment variables.
type Config struct {
	PostgresURL      string `env:"POSTGRES_URL,required"`
	RedisURL         string `env:"REDIS_URL,required"`
	JWTSecret        string `env:"JWT_SECRET,required"`
	SimulatorEnabled bool   `env:"SIMULATOR_ENABLED" envDefault:"true"`
}

// Load reads configuration from a .env file and environment variables.
func Load() (*Config, error) {
	// Load .env file, but it's okay if it doesn't exist
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
