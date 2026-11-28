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

func TestPostgresTextArrayValueEncodesUUIDAndEmptyArrays(t *testing.T) {
	tests := []struct {
		name   string
		values postgresTextArray
		want   string
	}{
		{
			name:   "uuid is encoded as an array literal",
			values: postgresTextArray{"24f3cb06-f409-4b3b-b23c-8d5a03540b5a"},
			want:   `{"24f3cb06-f409-4b3b-b23c-8d5a03540b5a"}`,
		},
		{name: "empty array", values: postgresTextArray{}, want: `{}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := test.values.Value()
			if err != nil {
				t.Fatalf("encode text array: %v", err)
			}
			if got, ok := value.(string); !ok || got != test.want {
				t.Fatalf("array value = %#v, want %q", value, test.want)
			}
		})
	}
}
