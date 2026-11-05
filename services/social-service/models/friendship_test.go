package models

import (
	"encoding/json"
	"testing"
)

func TestFriendshipIDJSONAlwaysUsesStrings(t *testing.T) {
	tests := []struct {
		name string
		id   FriendshipID
		want string
	}{
		{name: "legacy numeric ID", id: "123", want: `{"id_amistad":"123"}`},
		{name: "UUID ID", id: "d2719cd6-3a4f-4f4a-9c2a-0b52f4ca8b21", want: `{"id_amistad":"d2719cd6-3a4f-4f4a-9c2a-0b52f4ca8b21"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(struct {
				ID FriendshipID `json:"id_amistad"`
			}{ID: test.id})
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(data) != test.want {
				t.Fatalf("Marshal() = %s, want %s", data, test.want)
			}

			var decoded struct {
				ID FriendshipID `json:"id_amistad"`
			}
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if decoded.ID != test.id {
				t.Fatalf("Unmarshal() ID = %q, want %q", decoded.ID, test.id)
			}
		})
	}
}

func TestFriendshipIDUnmarshalAcceptsLegacyNumericJSON(t *testing.T) {
	var decoded struct {
		ID FriendshipID `json:"id_amistad"`
	}
	if err := json.Unmarshal([]byte(`{"id_amistad":123}`), &decoded); err != nil {
		t.Fatalf("Unmarshal() legacy numeric ID error = %v", err)
	}
	if decoded.ID != "123" {
		t.Fatalf("Unmarshal() ID = %q, want %q", decoded.ID, "123")
	}

	data, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("Marshal() legacy ID error = %v", err)
	}
	if string(data) != `{"id_amistad":"123"}` {
		t.Fatalf("Marshal() legacy ID = %s, want %s", data, `{"id_amistad":"123"}`)
	}
}

func TestFriendshipIDScan(t *testing.T) {
	for _, value := range []any{"d2719cd6-3a4f-4f4a-9c2a-0b52f4ca8b21", []byte("123")} {
		var id FriendshipID
		if err := id.Scan(value); err != nil {
			t.Fatalf("Scan(%T) error = %v", value, err)
		}
		if id == "" {
			t.Fatalf("Scan(%T) produced an empty ID", value)
		}
	}
}
