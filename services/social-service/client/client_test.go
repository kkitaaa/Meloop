package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meloop/services/common/httpresponse"
	"github.com/meloop/social-service/client"
	"github.com/meloop/social-service/models"
)

func TestSocialClient_ExcludeBlocked(t *testing.T) {
	ctx := context.Background()

	t.Run("Exitoso con exclusión y candidatos conservados", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/internal/blocks/exclude" {
				http.NotFound(w, r)
				return
			}
			var req models.ExcludeBlockedRequest
			_ = json.NewDecoder(r.Body).Decode(&req)

			httpresponse.Success(w, http.StatusOK, models.ExcludeBlockedResponse{
				CandidateIDs: []string{"cand-clean-1", "cand-clean-2"},
				ExcludedIDs:  []string{"cand-blocked"},
			})
		}))
		defer server.Close()

		cli := client.NewSocialClient(server.URL)
		allowed, err := cli.ExcludeBlocked(ctx, "usr-ref", []string{"cand-clean-1", "cand-blocked", "cand-clean-2"})
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		if len(allowed) != 2 || allowed[0] != "cand-clean-1" || allowed[1] != "cand-clean-2" {
			t.Errorf("candidatos permitidos inesperados: %v", allowed)
		}
	})

	t.Run("Obtiene ambos conjuntos: permitidos y excluidos", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpresponse.Success(w, http.StatusOK, models.ExcludeBlockedResponse{
				CandidateIDs: []string{"cand-clean"},
				ExcludedIDs:  []string{"cand-blocked-1", "cand-blocked-2"},
			})
		}))
		defer server.Close()

		cli := client.NewSocialClient(server.URL)
		allowed, excluded, err := cli.GetExcludedAndAllowed(ctx, "usr-ref", []string{"cand-clean", "cand-blocked-1", "cand-blocked-2"})
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}

		if len(allowed) != 1 || allowed[0] != "cand-clean" {
			t.Errorf("candidatos permitidos inesperados: %v", allowed)
		}
		if len(excluded) != 2 || excluded[0] != "cand-blocked-1" {
			t.Errorf("candidatos excluidos inesperados: %v", excluded)
		}
	})

	t.Run("Lista vacía retorna slices vacíos sin petición HTTP", func(t *testing.T) {
		cli := client.NewSocialClient("http://invalid-url-that-would-fail")
		allowed, err := cli.ExcludeBlocked(ctx, "usr-ref", []string{})
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if allowed == nil || len(allowed) != 0 {
			t.Errorf("se esperaba slice vacío")
		}
	})

	t.Run("Usuario de referencia vacío retorna error de validación", func(t *testing.T) {
		cli := client.NewSocialClient("http://any-url")
		_, err := cli.ExcludeBlocked(ctx, "   ", []string{"cand-1"})
		if err == nil {
			t.Errorf("se esperaba error con usuario vacío")
		}
	})

	t.Run("Error retornado por el servidor", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpresponse.Error(w, http.StatusBadRequest, "BAD_REQUEST", "Datos inválidos")
		}))
		defer server.Close()

		cli := client.NewSocialClient(server.URL)
		_, err := cli.ExcludeBlocked(ctx, "usr-ref", []string{"cand-1"})
		if err == nil {
			t.Errorf("se esperaba error del servidor")
		}
	})
}
