package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/untibullet/secure-db-manager/internal/config"
	"github.com/untibullet/secure-db-manager/internal/domain"
	"github.com/untibullet/secure-db-manager/internal/service"
	"github.com/untibullet/secure-db-manager/internal/session"
)

// MockAuthRepo implements service.AuthRepo.
type MockAuthRepo struct{ mock.Mock }

func (m *MockAuthRepo) GetUserForLogin(ctx context.Context, username string) (*domain.UserAuth, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserAuth), args.Error(1)
}

// testBcryptHash is a real bcrypt hash of "correct", computed once at package init.
var testBcryptHash string

func init() {
	h, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	testBcryptHash = string(h)
}

func newTestAuthService() *service.AuthService {
	return service.NewAuthService(
		session.NewStore(),
		&config.Config{
			DBHost:     "127.0.0.1",
			DBPort:     "5999", // no PG here → pgx.Connect will fail
			DBName:     "noexist",
			JWTSecret:  "test-secret",
			SessionTTL: time.Hour,
		},
	)
}

func TestAuthService_Login(t *testing.T) {
	svc := newTestAuthService()

	tests := []struct {
		name     string
		setup    func(*MockAuthRepo)
		password string
		wantErr  error
		timeout  time.Duration
	}{
		{
			name: "user_not_found",
			setup: func(r *MockAuthRepo) {
				r.On("GetUserForLogin", mock.Anything, "alice").Return(nil, domain.ErrNotFound)
			},
			password: "any",
			wantErr:  domain.ErrUnauthorized,
		},
		{
			name: "user_inactive",
			setup: func(r *MockAuthRepo) {
				r.On("GetUserForLogin", mock.Anything, "alice").Return(
					&domain.UserAuth{IsActive: false}, nil,
				)
			},
			password: "any",
			wantErr:  domain.ErrForbidden,
		},
		{
			name: "account_locked",
			setup: func(r *MockAuthRepo) {
				future := time.Now().Add(time.Hour)
				r.On("GetUserForLogin", mock.Anything, "alice").Return(
					&domain.UserAuth{IsActive: true, AccountLockedUntil: &future}, nil,
				)
			},
			password: "any",
			wantErr:  domain.ErrForbidden,
		},
		{
			name: "wrong_password",
			setup: func(r *MockAuthRepo) {
				r.On("GetUserForLogin", mock.Anything, "alice").Return(
					&domain.UserAuth{IsActive: true, PasswordHash: testBcryptHash}, nil,
				)
			},
			password: "wrong",
			wantErr:  domain.ErrUnauthorized,
		},
		{
			name: "pgx_connect_fails",
			setup: func(r *MockAuthRepo) {
				r.On("GetUserForLogin", mock.Anything, "alice").Return(
					&domain.UserAuth{
						UserID:       1,
						Username:     "alice",
						IsActive:     true,
						PasswordHash: testBcryptHash,
						RoleCode:     "TESTER",
					},
					nil,
				)
			},
			password: "correct",
			wantErr:  domain.ErrUnauthorized,
			timeout:  3 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAuthRepo)
			tt.setup(repo)

			ctx := context.Background()
			if tt.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.timeout)
				defer cancel()
			}

			_, err := svc.Login(ctx, repo, "alice", tt.password)
			require.ErrorIs(t, err, tt.wantErr)
			repo.AssertExpectations(t)
		})
	}
}
