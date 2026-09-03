package config

import (
	"errors"
	"fmt"
	"os"
)

var ErrMissingEnv = errors.New("config: missing required environment variable")

type Config struct {
	ConnectionString string
	JWTSecret        string
}

func Load() (*Config, error) {
	connString, ok := os.LookupEnv("DATABASE_URL")
	if !ok || connString == "" {
		return nil, fmt.Errorf("%w: DATABASE_URL", ErrMissingEnv)
	}

	jwtSecret, ok := os.LookupEnv("JWT_SECRET")
	if !ok || jwtSecret == "" {
		return nil, fmt.Errorf("%w: JWT_SECRET", ErrMissingEnv)
	}

	return &Config{
		ConnectionString: connString,
		JWTSecret:        jwtSecret,
	}, nil
}
