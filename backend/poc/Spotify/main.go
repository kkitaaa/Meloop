package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type SpotifySearchResponse struct {
	Artists struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			URI  string `json:"uri"`

			ExternalURLs struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
		} `json:"items"`
	} `json:"artists"`
}

func getAccessToken(clientID, clientSecret string) (string, error) {
	tokenURL := "https://accounts.spotify.com/api/token"

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(
		http.MethodPost,
		tokenURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", err
	}

	credentials := clientID + ":" + clientSecret
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))

	req.Header.Set("Authorization", "Basic "+encodedCredentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"Spotify devolvió HTTP %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var tokenResponse TokenResponse

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func searchArtist(accessToken, artistName string) (*SpotifySearchResponse, error) {
	searchURL := "https://api.spotify.com/v1/search"

	params := url.Values{}
	params.Set("q", artistName)
	params.Set("type", "artist")
	params.Set("limit", "1")

	fullURL := searchURL + "?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"Spotify devolvió HTTP %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var result SpotifySearchResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Advertencia: no se encontró archivo .env")
	}

	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		fmt.Println("Error: faltan SPOTIFY_CLIENT_ID o SPOTIFY_CLIENT_SECRET")
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Uso:")
		fmt.Println(`go run . "Nombre del artista"`)
		return
	}

	artistName := strings.Join(os.Args[1:], " ")

	fmt.Println("Buscando artista:", artistName)
	fmt.Println()

	accessToken, err := getAccessToken(clientID, clientSecret)
	if err != nil {
		fmt.Println("Error obteniendo token:", err)
		return
	}

	result, err := searchArtist(accessToken, artistName)
	if err != nil {
		fmt.Println("Error buscando artista:", err)
		return
	}

	if len(result.Artists.Items) == 0 {
		fmt.Println("No se encontraron artistas.")
		return
	}

	artist := result.Artists.Items[0]

	fmt.Println("Artista encontrado")
	fmt.Println("-------------------")
	fmt.Println("Nombre:", artist.Name)
	fmt.Println("Spotify ID:", artist.ID)
	fmt.Println("Spotify URI:", artist.URI)
	fmt.Println("Spotify URL:", artist.ExternalURLs.Spotify)
}
