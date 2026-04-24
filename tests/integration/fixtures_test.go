//go:build integration

package integration_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectAs opens a connection to the test container as the given user.
// The connection is automatically closed when t finishes.
func connectAs(t *testing.T, username, password string) *pgx.Conn {
	t.Helper()
	cfg, err := pgx.ParseConfig(containerDSN)
	require.NoError(t, err)
	cfg.User = username
	cfg.Password = password
	conn, err := pgx.ConnectConfig(context.Background(), cfg)
	require.NoError(t, err, "connect as %s", username)
	t.Cleanup(func() { conn.Close(context.Background()) }) //nolint:errcheck
	return conn
}

// execAs executes sql as the given connection, expecting success.
func execAs(t *testing.T, conn *pgx.Conn, sql string, args ...any) {
	t.Helper()
	_, err := conn.Exec(context.Background(), sql, args...)
	require.NoError(t, err)
}

// mustFail executes sql and asserts it returns an error.
// Optionally checks that the error is a PostgreSQL privilege violation (42501).
func mustFail(t *testing.T, conn *pgx.Conn, sql string, args ...any) {
	t.Helper()
	_, err := conn.Exec(context.Background(), sql, args...)
	require.Error(t, err, "expected error for: %s", sql)

	// Surface specific PG error code if available, but don't require it.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		assert.Equal(t, "42501", pgErr.Code,
			fmt.Sprintf("expected 42501 (insufficient_privilege), got %s: %s",
				pgErr.Code, pgErr.Message))
	}
}

// collectIDs runs a query that returns a single int column and gathers all values.
func collectIDs(t *testing.T, conn *pgx.Conn, query string, args ...any) []int {
	t.Helper()
	rows, err := conn.Query(context.Background(), query, args...)
	require.NoError(t, err)
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	require.NoError(t, rows.Err())
	return ids
}
