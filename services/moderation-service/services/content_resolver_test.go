package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meloop/moderation-service/models"
)

func TestHTTPContentResolverResolvesPostAndCommentSnapshots(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/internal/moderation/targets/post/post-1" {
			t.Errorf("unexpected target path: %s", request.URL.Path)
		}
		if request.Header.Get("X-User-ID") != "reporter-1" {
			t.Errorf("reporter header = %q", request.Header.Get("X-User-ID"))
		}
		if request.Header.Get("X-Internal-Service-Token") != "internal-token" {
			t.Errorf("missing internal service token")
		}
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"author_id":   "author-2",
			"author_name": "Artist",
			"text":        "hello",
			"urls":        []string{"https://example.test/image"},
		})
	}))
	defer server.Close()

	resolver := NewHTTPContentResolver(server.URL, "", "internal-token", server.Client())
	target, err := resolver.Resolve(context.Background(), "reporter-1", models.TargetPost, "post-1")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if target.AuthorID != "author-2" || target.Snapshot.Text != "hello" || target.Snapshot.AuthorName != "Artist" {
		t.Fatalf("unexpected target: %+v", target)
	}
	if len(target.Snapshot.URLs) != 1 || target.Snapshot.URLs[0] != "https://example.test/image" {
		t.Fatalf("unexpected snapshot URLs: %#v", target.Snapshot.URLs)
	}
}

func TestHTTPContentResolverResolvesVisibleUserProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/users/user-2/profile" {
			t.Errorf("unexpected user profile path: %s", request.URL.Path)
		}
		if request.Header.Get("X-User-ID") != "reporter-1" {
			t.Errorf("reporter header = %q", request.Header.Get("X-User-ID"))
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"success":true,"data":{"id_usuario":"user-2","username":"Melo","biografia":"bio https://example.test/docs.","foto_perfil":"https://example.test/avatar","banner":""}}`))
	}))
	defer server.Close()

	resolver := NewHTTPContentResolver("", server.URL, "", server.Client())
	target, err := resolver.Resolve(context.Background(), "reporter-1", models.TargetUser, "user-2")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if target.AuthorID != "user-2" || target.Snapshot.Text != "Melo bio https://example.test/docs." || target.Snapshot.AuthorName != "Melo" {
		t.Fatalf("unexpected target: %+v", target)
	}
	if len(target.Snapshot.URLs) != 2 || target.Snapshot.URLs[0] != "https://example.test/avatar" || target.Snapshot.URLs[1] != "https://example.test/docs" {
		t.Fatalf("unexpected snapshot URLs: %#v", target.Snapshot.URLs)
	}
}

func TestHTTPContentResolverMapsInvisibleTargetsToNotFound(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	resolver := NewHTTPContentResolver(server.URL, "", "", server.Client())
	_, err := resolver.Resolve(context.Background(), "reporter-1", models.TargetComment, "comment-1")
	if !errors.Is(err, ErrTargetNotFound) {
		t.Fatalf("Resolve() error = %v, want ErrTargetNotFound", err)
	}
}
