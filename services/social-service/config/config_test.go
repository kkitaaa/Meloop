package config

import "testing"

func TestLoadAuthServiceURL(t *testing.T) {
	t.Setenv("AUTH_SERVICE_URL", "")
	if got := Load().AuthServiceURL; got != "http://localhost:8083" {
		t.Fatalf("AuthServiceURL = %q, want http://localhost:8083", got)
	}

	t.Setenv("AUTH_SERVICE_URL", "http://auth-service:8083/")
	if got := Load().AuthServiceURL; got != "http://auth-service:8083/" {
		t.Fatalf("AuthServiceURL = %q, want configured URL", got)
	}
}
