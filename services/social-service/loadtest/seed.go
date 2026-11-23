package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	userCount           = 500
	virtualUserCount    = 200
	friendsPerUser      = 30
	neighborsPerSide    = friendsPerUser / 2
	loadTestDatabase    = "meloop_social_loadtest"
	loadTestDatabaseDSN = "postgres://postgres@127.0.0.1:15432/meloop_social_loadtest?sslmode=disable"
	loadTestAuthURL     = "http://127.0.0.1:18083"
)

type friendshipRow struct {
	id      string
	userID1 string
	userID2 string
}

type preferenceRow struct {
	id      string
	userID  string
	kind    string
	spotify string
}

type loadTestUser struct {
	UserID                    string   `json:"user_id"`
	Token                     string   `json:"token"`
	ExpectedFriendIDs         []string `json:"expected_friend_ids"`
	ExpectedSuggestionID      string   `json:"expected_suggestion_id"`
	ExpectedMutualFriendCount int      `json:"expected_mutual_friends"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "load-test seed failed:", err)
		os.Exit(1)
	}
}

func run() error {
	password := os.Getenv("LOADTEST_USER_PASSWORD")
	if len(password) < 8 || len(password) > 72 {
		return errors.New("LOADTEST_USER_PASSWORD must contain 8 to 72 bytes")
	}

	tokenFile := os.Getenv("LOADTEST_TOKEN_FILE")
	if tokenFile == "" {
		return errors.New("LOADTEST_TOKEN_FILE must point outside the repository")
	}
	tokenFile, err := filepath.Abs(tokenFile)
	if err != nil {
		return fmt.Errorf("resolve token file path: %w", err)
	}
	if err := ensureOutsideRepository(tokenFile); err != nil {
		return err
	}

	databaseURL := os.Getenv("LOADTEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = loadTestDatabaseDSN
	}
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("parse load-test database URL: %w", err)
	}
	if poolConfig.ConnConfig.Host != "127.0.0.1" ||
		poolConfig.ConnConfig.Port != 15432 ||
		poolConfig.ConnConfig.Database != loadTestDatabase {
		return errors.New("refusing to seed: expected the dedicated local load-test PostgreSQL database on 127.0.0.1:15432")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("open load-test database: %w", err)
	}
	defer pool.Close()
	if err := pingDatabase(ctx, pool); err != nil {
		return err
	}
	if err := verifySchema(ctx, pool); err != nil {
		return err
	}

	authBaseURL := os.Getenv("LOADTEST_AUTH_URL")
	if authBaseURL == "" {
		authBaseURL = loadTestAuthURL
	}
	authURL, err := url.Parse(authBaseURL)
	if err != nil || authURL.Scheme != "http" || authURL.Host != "127.0.0.1:18083" {
		return errors.New("refusing to authenticate outside the dedicated local load-test Auth Service on 127.0.0.1:18083")
	}

	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash load-test password: %w", err)
	}
	passwordHash := string(passwordHashBytes)

	client := &http.Client{Timeout: 20 * time.Second}
	userIDs := make([]string, userCount)
	loadTestUsers := make([]loadTestUser, virtualUserCount)
	for i := 0; i < userCount; i++ {
		username := fmt.Sprintf("lt_social_%03d", i)
		email := fmt.Sprintf("lt_social_%03d@loadtest.invalid", i)
		userID, err := registerOrFindUser(ctx, client, authBaseURL, pool, username, email, password, passwordHash)
		if err != nil {
			return fmt.Errorf("prepare test user %03d: %w", i, err)
		}
		userIDs[i] = userID

		if i < virtualUserCount {
			token, authenticatedID, err := login(ctx, client, authBaseURL, email, password)
			if err != nil {
				return fmt.Errorf("create session for virtual user %03d: %w", i, err)
			}
			if authenticatedID != userID {
				return fmt.Errorf("Auth Service returned a different identity for virtual user %03d", i)
			}
			loadTestUsers[i] = loadTestUser{
				UserID: userID,
				Token:  token,
			}
		}
	}

	for i := range loadTestUsers {
		loadTestUsers[i].ExpectedFriendIDs = cycleFriendIDs(userIDs, i)
		loadTestUsers[i].ExpectedSuggestionID = userIDs[(i+neighborsPerSide+1)%userCount]
		loadTestUsers[i].ExpectedMutualFriendCount = neighborsPerSide
	}
	uniqueIDs := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if _, exists := uniqueIDs[userID]; exists {
			return errors.New("Auth Service returned a duplicate identity for the load-test accounts")
		}
		uniqueIDs[userID] = struct{}{}
	}

	edges := buildEdges(userIDs)
	preferences := buildPreferences(userIDs)
	if err := seedData(ctx, pool, userIDs, edges, preferences); err != nil {
		return err
	}

	encodedUsers, err := json.Marshal(loadTestUsers)
	if err != nil {
		return fmt.Errorf("encode test sessions: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(tokenFile), 0o700); err != nil {
		return fmt.Errorf("create token directory: %w", err)
	}
	if err := os.WriteFile(tokenFile, encodedUsers, 0o600); err != nil {
		return fmt.Errorf("write test sessions outside repository: %w", err)
	}

	fmt.Printf("Verified %d users, %d accepted friendships, %d musical preferences, and %d valid sessions.\n",
		userCount, len(edges), len(preferences), len(loadTestUsers))
	return nil
}

func ensureOutsideRepository(tokenFile string) error {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}
	repositoryRoot, err := filepath.Abs(filepath.Join(workingDirectory, "..", ".."))
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	relativePath, err := filepath.Rel(repositoryRoot, tokenFile)
	if err != nil {
		return fmt.Errorf("check token file location: %w", err)
	}
	if relativePath == "." || (relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator))) {
		return errors.New("refusing to write session tokens inside the repository")
	}
	return nil
}

func pingDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	for attempt := 0; attempt < 30; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := pool.Ping(pingCtx)
		cancel()
		if err == nil {
			var databaseName string
			if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
				return fmt.Errorf("verify database identity: %w", err)
			}
			if databaseName != loadTestDatabase {
				return errors.New("refusing to seed a database other than the dedicated load-test database")
			}
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("dedicated load-test PostgreSQL did not become ready")
}

func verifySchema(ctx context.Context, pool *pgxpool.Pool) error {
	var hasUsers, hasFriendships, hasPreferences bool
	err := pool.QueryRow(ctx, `
		SELECT to_regclass('public.usuario') IS NOT NULL,
		       to_regclass('public.amistad') IS NOT NULL,
		       to_regclass('public.preferencia_musical') IS NOT NULL`,
	).Scan(&hasUsers, &hasFriendships, &hasPreferences)
	if err != nil {
		return fmt.Errorf("verify migrated schema: %w", err)
	}
	if !hasUsers || !hasFriendships || !hasPreferences {
		return errors.New("load-test migrations did not create the required social tables")
	}
	return nil
}

func registerOrFindUser(
	ctx context.Context,
	client *http.Client,
	authBaseURL string,
	pool *pgxpool.Pool,
	username, email, password, passwordHash string,
) (string, error) {
	payload, err := json.Marshal(map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	})
	if err != nil {
		return "", err
	}

	status, body, err := postJSON(ctx, client, authBaseURL+"/auth/register", payload)
	if err != nil {
		return "", err
	}
	if status == http.StatusCreated {
		var response struct {
			Success bool `json:"success"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &response); err != nil || !response.Success || response.Data.ID == "" {
			return "", errors.New("Auth Service returned an invalid registration response")
		}
		return response.Data.ID, nil
	}
	if status != http.StatusConflict {
		return "", fmt.Errorf("Auth Service registration returned HTTP %d", status)
	}

	var userID string
	err = pool.QueryRow(ctx,
		`UPDATE usuario SET contrasena_hash = $1 WHERE username = $2 AND correo = $3 RETURNING id_usuario::text`,
		passwordHash, username, email,
	).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("find existing dedicated load-test account: %w", err)
	}
	return userID, nil
}

func login(
	ctx context.Context,
	client *http.Client,
	authBaseURL, email, password string,
) (string, string, error) {
	payload, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return "", "", err
	}

	status, body, err := postJSON(ctx, client, authBaseURL+"/auth/login", payload)
	if err != nil {
		return "", "", err
	}
	if status != http.StatusOK {
		return "", "", fmt.Errorf("Auth Service login returned HTTP %d", status)
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			SessionToken string `json:"session_token"`
			User         struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil || !response.Success ||
		response.Data.SessionToken == "" || response.Data.User.ID == "" {
		return "", "", errors.New("Auth Service returned an invalid login response")
	}
	return response.Data.SessionToken, response.Data.User.ID, nil
}

func postJSON(ctx context.Context, client *http.Client, target string, payload []byte) (int, []byte, error) {
	for attempt := 0; attempt < 10; attempt++ {
		requestCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		request, err := http.NewRequestWithContext(requestCtx, http.MethodPost, target, bytes.NewReader(payload))
		if err != nil {
			cancel()
			return 0, nil, errors.New("create Auth Service request")
		}
		request.Header.Set("Content-Type", "application/json")
		response, err := client.Do(request)
		if err != nil {
			cancel()
			var netErr net.Error
			if errors.As(err, &netErr) && attempt < 9 {
				time.Sleep(time.Second)
				continue
			}
			return 0, nil, errors.New("Auth Service request failed")
		}

		body, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		closeErr := response.Body.Close()
		cancel()
		if readErr != nil {
			return 0, nil, errors.New("read Auth Service response")
		}
		if closeErr != nil {
			return 0, nil, errors.New("close Auth Service response")
		}
		return response.StatusCode, body, nil
	}
	return 0, nil, errors.New("Auth Service did not become reachable")
}

func cycleFriendIDs(userIDs []string, userIndex int) []string {
	friends := make([]string, 0, friendsPerUser)
	for offset := 1; offset <= neighborsPerSide; offset++ {
		friends = append(friends, userIDs[(userIndex+offset)%len(userIDs)])
		friends = append(friends, userIDs[(userIndex-offset+len(userIDs))%len(userIDs)])
	}
	return friends
}

func buildEdges(userIDs []string) []friendshipRow {
	edges := make([]friendshipRow, 0, len(userIDs)*neighborsPerSide)
	for i, userID := range userIDs {
		for offset := 1; offset <= neighborsPerSide; offset++ {
			neighborIndex := (i + offset) % len(userIDs)
			edges = append(edges, friendshipRow{
				id:      fmt.Sprintf("lt_social_edge_%03d_%03d", i, neighborIndex),
				userID1: userID,
				userID2: userIDs[neighborIndex],
			})
		}
	}
	return edges
}

func buildPreferences(userIDs []string) []preferenceRow {
	preferences := make([]preferenceRow, 0, len(userIDs)*10)
	for i, userID := range userIDs {
		for offset := 0; offset < 2; offset++ {
			preferences = append(preferences, preferenceRow{
				id:      fmt.Sprintf("lt_social_pref_%03d_genre_%d", i, offset),
				userID:  userID,
				kind:    "GENERO",
				spotify: fmt.Sprintf("lt_genre_%02d", (i+offset*3)%20),
			})
		}
		for offset := 0; offset < 3; offset++ {
			preferences = append(preferences, preferenceRow{
				id:      fmt.Sprintf("lt_social_pref_%03d_artist_%d", i, offset),
				userID:  userID,
				kind:    "ARTISTA",
				spotify: fmt.Sprintf("lt_artist_%03d", (i*3+offset)%100),
			})
		}
		for offset := 0; offset < 5; offset++ {
			preferences = append(preferences, preferenceRow{
				id:      fmt.Sprintf("lt_social_pref_%03d_track_%d", i, offset),
				userID:  userID,
				kind:    "CANCION",
				spotify: fmt.Sprintf("lt_track_%03d", (i*5+offset)%250),
			})
		}
	}
	return preferences
}

func seedData(
	ctx context.Context,
	pool *pgxpool.Pool,
	userIDs []string,
	edges []friendshipRow,
	preferences []preferenceRow,
) error {
	edgeIDs := make([]string, len(edges))
	edgeRows := make([][]any, len(edges))
	for i, edge := range edges {
		edgeIDs[i] = edge.id
		edgeRows[i] = []any{edge.id, edge.userID1, edge.userID2, "ACEPTADA"}
	}
	preferenceIDs := make([]string, len(preferences))
	preferenceRows := make([][]any, len(preferences))
	for i, preference := range preferences {
		preferenceIDs[i] = preference.id
		preferenceRows[i] = []any{
			preference.id, preference.userID, preference.kind, preference.spotify,
		}
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin fixture transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM amistad WHERE id_amistad = ANY($1)`, edgeIDs); err != nil {
		return fmt.Errorf("replace only load-test friendships: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM preferencia_musical WHERE id_preferencia = ANY($1)`, preferenceIDs); err != nil {
		return fmt.Errorf("replace only load-test preferences: %w", err)
	}

	edgeCount, err := tx.CopyFrom(ctx,
		pgx.Identifier{"public", "amistad"},
		[]string{"id_amistad", "id_usuario_1", "id_usuario_2", "estado"},
		pgx.CopyFromRows(edgeRows),
	)
	if err != nil {
		return fmt.Errorf("insert load-test friendships: %w", err)
	}
	preferenceCount, err := tx.CopyFrom(ctx,
		pgx.Identifier{"public", "preferencia_musical"},
		[]string{"id_preferencia", "id_usuario", "tipo", "spotify_id"},
		pgx.CopyFromRows(preferenceRows),
	)
	if err != nil {
		return fmt.Errorf("insert load-test musical preferences: %w", err)
	}
	if edgeCount != int64(len(edges)) || preferenceCount != int64(len(preferences)) {
		return errors.New("database inserted an unexpected number of fixture rows")
	}

	var userRowCount, acceptedCount, preferenceRowCount, selfFriendships, duplicatePairs int
	var degreeCount, invalidDegrees int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM usuario WHERE id_usuario = ANY($1)`, userIDs,
	).Scan(&userRowCount); err != nil {
		return fmt.Errorf("verify load-test user count: %w", err)
	}
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM amistad
		WHERE id_amistad = ANY($1) AND estado = 'ACEPTADA'`, edgeIDs,
	).Scan(&acceptedCount); err != nil {
		return fmt.Errorf("verify accepted friendship count: %w", err)
	}
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM preferencia_musical
		WHERE id_preferencia = ANY($1)`, preferenceIDs,
	).Scan(&preferenceRowCount); err != nil {
		return fmt.Errorf("verify musical preference count: %w", err)
	}
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM amistad
		WHERE id_amistad = ANY($1) AND id_usuario_1 = id_usuario_2`, edgeIDs,
	).Scan(&selfFriendships); err != nil {
		return fmt.Errorf("verify self-friendships: %w", err)
	}
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT LEAST(id_usuario_1, id_usuario_2), GREATEST(id_usuario_1, id_usuario_2)
			FROM amistad
			WHERE id_amistad = ANY($1)
			GROUP BY 1, 2
			HAVING COUNT(*) > 1
		) duplicate_pairs`, edgeIDs,
	).Scan(&duplicatePairs); err != nil {
		return fmt.Errorf("verify duplicate friendship pairs: %w", err)
	}
	if err := tx.QueryRow(ctx, `
		WITH fixture_friendships AS (
			SELECT id_usuario_1 AS user_id, id_usuario_2 AS friend_id
			FROM amistad WHERE id_amistad = ANY($1)
			UNION ALL
			SELECT id_usuario_2 AS user_id, id_usuario_1 AS friend_id
			FROM amistad WHERE id_amistad = ANY($1)
		), degrees AS (
			SELECT user_id, COUNT(*) AS degree
			FROM fixture_friendships
			GROUP BY user_id
		)
		SELECT COUNT(*), COUNT(*) FILTER (WHERE degree <> $2) FROM degrees`,
		edgeIDs, friendsPerUser,
	).Scan(&degreeCount, &invalidDegrees); err != nil {
		return fmt.Errorf("verify friendship degrees: %w", err)
	}
	if userRowCount != userCount || acceptedCount != len(edges) ||
		preferenceRowCount != len(preferences) || selfFriendships != 0 ||
		duplicatePairs != 0 || degreeCount != userCount || invalidDegrees != 0 {
		return fmt.Errorf("fixture integrity check failed (users=%d, accepted=%d, preferences=%d, self=%d, duplicate pairs=%d, represented users=%d, invalid degrees=%d)",
			userRowCount, acceptedCount, preferenceRowCount, selfFriendships, duplicatePairs, degreeCount, invalidDegrees)
	}
	if len(userIDs) != userCount || len(edges) != userCount*neighborsPerSide ||
		len(preferences) != userCount*10 {
		return errors.New("fixture generator produced unexpected totals")
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit fixture transaction: %w", err)
	}
	return nil
}
