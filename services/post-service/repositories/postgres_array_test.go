package repositories

import "testing"

func TestPostgresTextArrayValueQuotesAndEscapesValues(t *testing.T) {
	value, err := (postgresTextArray{`user,"one`, `user\ two`}).Value()
	if err != nil {
		t.Fatalf("encode text array: %v", err)
	}
	if got, want := value, `{"user,\"one","user\\ two"}`; got != want {
		t.Fatalf("array value = %q, want %q", got, want)
	}
}
