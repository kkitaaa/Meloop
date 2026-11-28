package repositories

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGeneratePresignedUploadURL_UsesPublicEndpoint(t *testing.T) {
	repository, err := NewMinIOStorageRepository(
		"minio:9000",
		"media-user",
		"media-password",
		"meloop-media",
		false,
		"us-east-1",
		"https://media.example.test",
	)
	if err != nil {
		t.Fatalf("no se pudo crear el repositorio MinIO: %v", err)
	}

	uploadURL, err := repository.GeneratePresignedUploadURL(
		context.Background(),
		"profiles/user-1/avatar.png",
		"image/png",
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("no se pudo generar la URL prefirmada: %v", err)
	}

	parsed, err := url.Parse(uploadURL)
	if err != nil {
		t.Fatalf("URL prefirmada inválida: %v", err)
	}
	if parsed.Host != "media.example.test" {
		t.Fatalf("la URL debe usar el host público, se obtuvo %q", parsed.Host)
	}
	if strings.Contains(parsed.Host, "minio:9000") {
		t.Fatalf("la URL expone el host interno de MinIO: %s", uploadURL)
	}

	downloadURL, err := repository.GeneratePresignedDownloadURL(
		context.Background(),
		"profiles/user-1/avatar.png",
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("no se pudo generar la URL prefirmada de descarga: %v", err)
	}
	download, err := url.Parse(downloadURL)
	if err != nil {
		t.Fatalf("URL prefirmada de descarga inválida: %v", err)
	}
	if download.Host != "media.example.test" {
		t.Fatalf("la URL de descarga debe usar el host público, se obtuvo %q", download.Host)
	}
}

func TestNewMinIOStorageRepository_RejectsInvalidPublicEndpoint(t *testing.T) {
	_, err := NewMinIOStorageRepository(
		"minio:9000",
		"media-user",
		"media-password",
		"meloop-media",
		false,
		"us-east-1",
		"minio:9000",
	)
	if err == nil {
		t.Fatal("se esperaba un error para un endpoint público sin esquema HTTP(S)")
	}
}
