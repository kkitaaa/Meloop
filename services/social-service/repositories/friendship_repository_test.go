package repositories

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgxpool"
)

type countQueryTimeoutTracer struct {
	deadline      time.Time
	deadlineFound chan struct{}
	cancelParent  context.CancelFunc
}

func (tracer *countQueryTimeoutTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if !strings.Contains(strings.ToLower(data.SQL), "count(*)") {
		return ctx
	}
	if deadline, ok := ctx.Deadline(); ok {
		tracer.deadline = deadline
	}
	if tracer.deadlineFound != nil {
		close(tracer.deadlineFound)
	}
	if tracer.cancelParent != nil {
		tracer.cancelParent()
	}
	return ctx
}

func (*countQueryTimeoutTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func newCountQueryTestPool(t *testing.T, tracer *countQueryTimeoutTracer) *pgxpool.Pool {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go serveCountQueryTestConnection(listener)

	config, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://test:test@%s/test?sslmode=disable", listener.Addr()))
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.Tracer = tracer
	config.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func serveCountQueryTestConnection(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	backend := pgproto3.NewBackend(conn, conn)
	if _, err := backend.ReceiveStartupMessage(); err != nil {
		return
	}
	backend.Send(&pgproto3.AuthenticationOk{})
	for _, status := range []pgproto3.ParameterStatus{
		{Name: "server_version", Value: "16.0"},
		{Name: "server_encoding", Value: "UTF8"},
		{Name: "client_encoding", Value: "UTF8"},
		{Name: "DateStyle", Value: "ISO, MDY"},
		{Name: "integer_datetimes", Value: "on"},
		{Name: "standard_conforming_strings", Value: "on"},
	} {
		backend.Send(&status)
	}
	backend.Send(&pgproto3.BackendKeyData{ProcessID: 1, SecretKey: []byte{0, 0, 0, 1}})
	backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
	if err := backend.Flush(); err != nil {
		return
	}

	for {
		message, err := backend.Receive()
		if err != nil {
			return
		}
		switch message := message.(type) {
		case *pgproto3.Parse:
			backend.Send(&pgproto3.ParseComplete{})
		case *pgproto3.Describe:
			if message.ObjectType == 'S' {
				backend.Send(&pgproto3.ParameterDescription{ParameterOIDs: []uint32{25}})
			}
			backend.Send(countQueryRowDescription())
		case *pgproto3.Bind:
			backend.Send(&pgproto3.BindComplete{})
		case *pgproto3.Execute:
			backend.Send(&pgproto3.DataRow{Values: [][]byte{make([]byte, 8)}})
			backend.Send(&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")})
		case *pgproto3.Sync:
			backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
			if err := backend.Flush(); err != nil {
				return
			}
		}
	}
}

func countQueryRowDescription() *pgproto3.RowDescription {
	return &pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{{
		Name:         []byte("count"),
		DataTypeOID:  20,
		DataTypeSize: 8,
		Format:       0,
	}}}
}

func TestFriendshipRepository_NilPool(t *testing.T) {
	repo := NewFriendshipRepository(nil)
	ctx := context.Background()

	if _, err := repo.CreatePending(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on CreatePending")
	}
	if _, err := repo.ListReceived(ctx, "u1"); err == nil {
		t.Error("expected error with nil pool on ListReceived")
	}
	if _, err := repo.ListSent(ctx, "u1"); err == nil {
		t.Error("expected error with nil pool on ListSent")
	}
	if _, err := repo.Accept(ctx, "1", "u1"); err == nil {
		t.Error("expected error with nil pool on Accept")
	}
	if _, err := repo.Reject(ctx, "1", "u1"); err == nil {
		t.Error("expected error with nil pool on Reject")
	}
	if _, err := repo.Cancel(ctx, "1", "u1"); err == nil {
		t.Error("expected error with nil pool on Cancel")
	}
	if _, err := repo.ListFriends(ctx, "u1", 20, "", nil); err == nil {
		t.Error("expected error with nil pool on ListFriends")
	}
	if err := repo.RemoveFriend(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on RemoveFriend")
	}
	if _, err := repo.GetFriendProfile(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on GetFriendProfile")
	}
	if _, err := repo.BlockUser(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on BlockUser")
	}
	if _, err := repo.IsBlocked(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on IsBlocked")
	}
	if _, err := repo.GetMutualFriendsQuery(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on GetMutualFriendsQuery")
	}
	if _, _, err := repo.GetMutualFriendsPage(ctx, "u1", "u2", "", 20); err == nil {
		t.Error("expected error with nil pool on GetMutualFriendsPage")
	}
	if _, err := repo.ListMutualFriendProfiles(ctx, []string{"u3"}); err == nil {
		t.Error("expected error with nil pool on ListMutualFriendProfiles")
	}
	if _, err := repo.GetEligibleCandidates(ctx, "u1"); err == nil {
		t.Error("expected error with nil pool on GetEligibleCandidates")
	}
	if _, err := repo.GetMutualFriendsCount(ctx, "u1", []string{"u2"}); err == nil {
		t.Error("expected error with nil pool on GetMutualFriendsCount")
	}
	if _, err := repo.ExcludeBlocked(ctx, "u1", []string{"u2"}); err == nil {
		t.Error("expected error with nil pool on ExcludeBlocked")
	}
	if _, err := repo.GetBlockedCandidateIDs(ctx, "u1", []string{"u2"}); err == nil {
		t.Error("expected error with nil pool on GetBlockedCandidateIDs")
	}
	if _, err := repo.CanSendMessage(ctx, "u1", "u2"); err == nil {
		t.Error("expected error with nil pool on CanSendMessage")
	}
}

func TestFriendshipRepository_CountQueriesUseThreeSecondTimeout(t *testing.T) {
	tests := []struct {
		name  string
		query func(*postgresFriendshipRepository, context.Context) error
	}{
		{
			name: "accepted friends",
			query: func(repo *postgresFriendshipRepository, ctx context.Context) error {
				_, err := repo.queryAcceptedFriendsCount(ctx, "u1")
				return err
			},
		},
		{
			name: "blocked users",
			query: func(repo *postgresFriendshipRepository, ctx context.Context) error {
				_, err := repo.queryBlockedUsersCount(ctx, "u1")
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tracer := &countQueryTimeoutTracer{deadlineFound: make(chan struct{})}
			repo := &postgresFriendshipRepository{pool: newCountQueryTestPool(t, tracer)}

			if err := test.query(repo, context.Background()); err != nil {
				t.Fatalf("count query returned an error: %v", err)
			}
			select {
			case <-tracer.deadlineFound:
			case <-time.After(time.Second):
				t.Fatal("count query did not start")
			}
			remaining := time.Until(tracer.deadline)
			if remaining > 3*time.Second || remaining < 2*time.Second {
				t.Fatalf("count query deadline has %v remaining, want approximately 3s", remaining)
			}
		})
	}
}

func TestFriendshipRepository_CountQueriesRespectParentCancellation(t *testing.T) {
	tests := []struct {
		name  string
		query func(*postgresFriendshipRepository, context.Context) error
	}{
		{
			name: "accepted friends",
			query: func(repo *postgresFriendshipRepository, ctx context.Context) error {
				_, err := repo.queryAcceptedFriendsCount(ctx, "u1")
				return err
			},
		},
		{
			name: "blocked users",
			query: func(repo *postgresFriendshipRepository, ctx context.Context) error {
				_, err := repo.queryBlockedUsersCount(ctx, "u1")
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			tracer := &countQueryTimeoutTracer{deadlineFound: make(chan struct{}), cancelParent: cancel}
			repo := &postgresFriendshipRepository{pool: newCountQueryTestPool(t, tracer)}

			err := test.query(repo, ctx)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("count query error = %v, want context.Canceled", err)
			}
			select {
			case <-tracer.deadlineFound:
			case <-time.After(time.Second):
				t.Fatal("count query did not start before parent cancellation")
			}
		})
	}
}
