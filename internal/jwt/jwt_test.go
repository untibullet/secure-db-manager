package jwt_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/untibullet/secure-db-manager/internal/jwt"
)

const testSecret = "test-secret"

func TestRoundTrip(t *testing.T) {
	userID := 42

	token, err := jwt.Generate(userID, testSecret, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := jwt.Parse(token, testSecret)
	require.NoError(t, err)

	got, err := claims.UserID()
	require.NoError(t, err)
	assert.Equal(t, userID, got)
}

func TestExpiredToken(t *testing.T) {
	token, err := jwt.Generate(1, testSecret, -time.Second)
	require.NoError(t, err)

	_, err = jwt.Parse(token, testSecret)
	assert.Error(t, err)
}

func TestWrongSecret(t *testing.T) {
	token, err := jwt.Generate(1, testSecret, time.Hour)
	require.NoError(t, err)

	_, err = jwt.Parse(token, "wrong-secret")
	assert.Error(t, err)
}
