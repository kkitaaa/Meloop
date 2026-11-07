package repositories

import (
	"regexp"
	"testing"
)

func TestNewNotificationConfigIDMatchesPostgresVarcharID(t *testing.T) {
	first, err := newNotificationConfigID()
	if err != nil {
		t.Fatalf("unexpected ID generation error: %v", err)
	}
	second, err := newNotificationConfigID()
	if err != nil {
		t.Fatalf("unexpected ID generation error: %v", err)
	}

	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidPattern.MatchString(first) {
		t.Fatalf("generated ID is not a UUID string: %q", first)
	}
	if first == second {
		t.Fatal("expected independently generated IDs to differ")
	}
}
