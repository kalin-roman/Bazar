package config

import (
	"errors"
	"fmt"
	"os"
)

var ErrMissingEnv = errors.New("config: missing required environment variable")

type Config struct {
	ConnectionString string
	JWKSURL          string
}

func Load() (*Config, error) {
	connString, ok := os.LookupEnv("DATABASE_URL")
	if !ok || connString == "" {
		return nil, fmt.Errorf("%w: DATABASE_URL", ErrMissingEnv)
	}

	jwksURL, ok := os.LookupEnv("JWKS_URL")
	if !ok || jwksURL == "" {
		return nil, fmt.Errorf("%w: JWKS_URL", ErrMissingEnv)
	}

	return &Config{
		ConnectionString: connString,
		JWKSURL:          jwksURL,
	}, nil
}
