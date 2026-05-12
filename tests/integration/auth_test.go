//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/untibullet/secure-db-manager/internal/config"
	"github.com/untibullet/secure-db-manager/internal/domain"
	"github.com/untibullet/secure-db-manager/internal/repository"
	"github.com/untibullet/secure-db-manager/internal/seclog"
	"github.com/untibullet/secure-db-manager/internal/service"
	"github.com/untibullet/secure-db-manager/internal/session"
)

// testConfig builds a Config pointing at the test container.
func testConfig() *config.Config {
	return &config.Config{
		DBHost:        containerHost,
		DBPort:        containerPort,
		DBName:        pgDB,
		SuperuserUser: pgUser,
		SuperuserPass: pgPassword,
		JWTSecret:     testJWTSecret,
		SessionTTL:    time.Hour,
	}
}

// TestAuth_Login_Success verifies that a valid username/password opens a PG connection
// and stores an active session (AD-11).
func TestAuth_Login_Success(t *testing.T) {
	store := session.NewStore()
	svc := service.NewAuthService(store, testConfig(), seclog.Noop())
	roleRepo := repository.NewRoleRepository(sysPool)

	token, err := svc.Login(context.Background(), roleRepo, "tester_a", "testerapass")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	sess, ok := store.Get(testerAUserID)
	require.True(t, ok, "session must exist after successful login")
	assert.NotNil(t, sess.Conn, "session must hold an open PG connection")

	t.Cleanup(func() { store.Delete(testerAUserID) })
}

// TestAuth_Login_WrongPassword verifies that bcrypt mismatch returns ErrUnauthorized
// without opening a PG connection (AD-11).
func TestAuth_Login_WrongPassword(t *testing.T) {
	store := session.NewStore()
	svc := service.NewAuthService(store, testConfig(), seclog.Noop())
	roleRepo := repository.NewRoleRepository(sysPool)

	_, err := svc.Login(context.Background(), roleRepo, "tester_a", "wrongpassword")
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	_, ok := store.Get(testerAUserID)
	assert.False(t, ok, "no session must be created on failed bcrypt check")
}

// TestAuth_Login_PGRoleDeleted verifies that a correct bcrypt hash but missing PG login
// role returns ErrUnauthorized (pgx.Connect failure path in AD-11).
func TestAuth_Login_PGRoleDeleted(t *testing.T) {
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("tmppass"), bcrypt.MinCost)
	require.NoError(t, err)

	_, err = superConn.Exec(ctx, "CREATE ROLE tmp_no_pg_role LOGIN PASSWORD 'tmppass'")
	require.NoError(t, err)

	var uid int
	err = superConn.QueryRow(ctx,
		`INSERT INTO users (username, password_hash, email, full_name, is_active)
		 VALUES ('tmp_no_pg_role', $1, 'tmp_no_pg_role@test.local', 'Tmp No PG Role', TRUE)
		 RETURNING user_id`,
		string(hash),
	).Scan(&uid)
	require.NoError(t, err)

	t.Cleanup(func() {
		superConn.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", uid) //nolint:errcheck
		superConn.Exec(context.Background(), "DROP ROLE IF EXISTS tmp_no_pg_role")         //nolint:errcheck
	})

	// Drop the PG role so pgx.Connect will fail even though bcrypt passes.
	_, err = superConn.Exec(ctx, "DROP ROLE tmp_no_pg_role")
	require.NoError(t, err)

	store := session.NewStore()
	svc := service.NewAuthService(store, testConfig(), seclog.Noop())
	roleRepo := repository.NewRoleRepository(sysPool)

	_, err = svc.Login(ctx, roleRepo, "tmp_no_pg_role", "tmppass")
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

// TestAuth_Login_ReplacesSession verifies that a second login for the same user
// replaces the existing session, closing the old connection (AD-8).
func TestAuth_Login_ReplacesSession(t *testing.T) {
	ctx := context.Background()
	store := session.NewStore()
	svc := service.NewAuthService(store, testConfig(), seclog.Noop())
	roleRepo := repository.NewRoleRepository(sysPool)

	_, err := svc.Login(ctx, roleRepo, "tester_b", "testerbpass")
	require.NoError(t, err)
	sess1, ok := store.Get(testerBUserID)
	require.True(t, ok)
	conn1 := sess1.Conn

	_, err = svc.Login(ctx, roleRepo, "tester_b", "testerbpass")
	require.NoError(t, err)
	sess2, ok := store.Get(testerBUserID)
	require.True(t, ok)

	assert.NotSame(t, conn1, sess2.Conn, "second login must create a new connection")

	t.Cleanup(func() { store.Delete(testerBUserID) })
}
