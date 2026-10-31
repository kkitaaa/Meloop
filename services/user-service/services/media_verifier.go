package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// ErrMediaNotFound indica que la referencia del objeto multimedia no existe en MinIO/S3
	ErrMediaNotFound = errors.New("MEDIA_NOT_FOUND")
	// ErrMediaStorageFailed indica que ocurrió un fallo de comunicación con media-service o el almacenamiento MinIO/S3
	ErrMediaStorageFailed = errors.New("MEDIA_STORAGE_FAILED")
)

// MediaVerifier define la interfaz para validar objetos multimedia a través de media-service
type MediaVerifier interface {
	VerifyMedia(ctx context.Context, objectKey string) (bool, error)
}

type httpMediaVerifier struct {
	mediaServiceURL string
	httpClient      *http.Client
}

// NewHTTPMediaVerifier crea un verificador HTTP que se comunica con media-service
func NewHTTPMediaVerifier(mediaServiceURL string) MediaVerifier {
	return &httpMediaVerifier{
		mediaServiceURL: strings.TrimRight(mediaServiceURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// VerifyMedia consulta a media-service para verificar si el archivo existe en MinIO/S3
func (v *httpMediaVerifier) VerifyMedia(ctx context.Context, objectKey string) (bool, error) {
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return true, nil
	}

	endpoint := fmt.Sprintf("%s/media/verify?key=%s", v.mediaServiceURL, url.QueryEscape(key))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("%w: error al crear petición a media-service: %v", ErrMediaStorageFailed, err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: fallo de conexión con media-service: %v", ErrMediaStorageFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, ErrMediaNotFound
	}

	// 502, 500 u otros errores indican fallo del almacenamiento de objetos
	return false, fmt.Errorf("%w: media-service devolvió estado %d", ErrMediaStorageFailed, resp.StatusCode)
}

// VerifyMediaWithFallback intenta POST y GET ante posibles variaciones de rutas
func (v *httpMediaVerifier) VerifyMediaWithFallback(ctx context.Context, objectKey string) (bool, error) {
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return true, nil
	}

	body, _ := json.Marshal(map[string]string{"object_key": key})
	endpoint := fmt.Sprintf("%s/media/verify", v.mediaServiceURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return v.VerifyMedia(ctx, key)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: fallo de conexión con media-service: %v", ErrMediaStorageFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, ErrMediaNotFound
	}

	return false, fmt.Errorf("%w: media-service devolvió estado %d", ErrMediaStorageFailed, resp.StatusCode)
}
