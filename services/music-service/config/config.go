package config

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

const defaultPort = "8091"

var marketPattern = regexp.MustCompile(`^[A-Za-z]{2}$`)

type Config struct {
	Port                string
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyMarket       string
	RedisURL            string
}

func Load() Config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = defaultPort
	}

	return Config{
		Port:                port,
		SpotifyClientID:     strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_ID")),
		SpotifyClientSecret: strings.TrimSpace(os.Getenv("SPOTIFY_CLIENT_SECRET")),
		SpotifyMarket:       strings.ToUpper(strings.TrimSpace(os.Getenv("SPOTIFY_MARKET"))),
		RedisURL:            strings.TrimSpace(os.Getenv("REDIS_URL")),
	}
}

func (c Config) Validate() error {
	if c.SpotifyClientID == "" {
		return errors.New("SPOTIFY_CLIENT_ID is required")
	}
	if c.SpotifyClientSecret == "" {
		return errors.New("SPOTIFY_CLIENT_SECRET is required")
	}
	if !marketPattern.MatchString(c.SpotifyMarket) {
		return errors.New("SPOTIFY_MARKET must be a two-letter ISO 3166-1 alpha-2 code")
	}
	return nil
}
