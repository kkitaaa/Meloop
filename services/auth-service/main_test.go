package main

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestLogConnectionResultDoesNotExposeCredentials(t *testing.T) {
	const (
		databaseURL = "postgres://test-user:db-test-secret@db.example.test/meloop?token=db-test-token"
		redisURL    = "redis://:redis-test-secret@redis.example.test:6379?token=redis-test-token"
	)

	tests := []struct {
		name      string
		component string
		err       error
		event     string
		result    string
		sensitive []string
	}{
		{
			name:      "database connection succeeds",
			component: "database",
			event:     "database_connected",
			result:    "connected",
			sensitive: []string{databaseURL},
		},
		{
			name:      "database connection fails",
			component: "database",
			err:       errors.New("connection failed: " + databaseURL),
			event:     "database_connection_failed",
			result:    "failed",
			sensitive: []string{databaseURL, "db-test-secret", "db-test-token"},
		},
		{
			name:      "redis connection succeeds",
			component: "redis",
			event:     "redis_connected",
			result:    "connected",
			sensitive: []string{redisURL},
		},
		{
			name:      "redis connection fails",
			component: "redis",
			err:       errors.New("connection failed: " + redisURL),
			event:     "redis_connection_failed",
			result:    "failed",
			sensitive: []string{redisURL, "redis-test-secret", "redis-test-token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&output, nil)).With("service", "auth-service")

			logConnectionResult(logger, tt.component, tt.err)

			log := output.String()
			for _, secret := range tt.sensitive {
				if strings.Contains(log, secret) {
					t.Errorf("log output contains sensitive connection data")
				}
			}
			for _, field := range []string{tt.event, `"service":"auth-service"`, `"component":"` + tt.component + `"`, `"result":"` + tt.result + `"`} {
				if !strings.Contains(log, field) {
					t.Errorf("log output is missing expected diagnostic information")
				}
			}
			if strings.Contains(log, `"error"`) || strings.Contains(log, `"url"`) {
				t.Errorf("log output contains an unfiltered error or connection URL field")
			}
		})
	}
}
