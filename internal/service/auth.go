package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	appjwt "github.com/untibullet/secure-db-manager/internal/jwt"
	"github.com/untibullet/secure-db-manager/internal/config"
	"github.com/untibullet/secure-db-manager/internal/domain"
	"github.com/untibullet/secure-db-manager/internal/logger"
	"github.com/untibullet/secure-db-manager/internal/session"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepo interface {
	GetUserForLogin(ctx context.Context, username string) (*domain.UserAuth, error)
}

type AuthService struct {
	store *session.Store
	cfg   *config.Config
}

func NewAuthService(store *session.Store, cfg *config.Config) *AuthService {
	return &AuthService{store: store, cfg: cfg}
}

// Login проверяет пароль через bcrypt, открывает pgx.Conn от имени пользователя,
// сохраняет сессию и возвращает JWT (AD-8, AD-11).
func (s *AuthService) Login(ctx context.Context, repo AuthRepo, username, password string) (string, error) {
	ua, err := repo.GetUserForLogin(ctx, username)
	if errors.Is(err, domain.ErrNotFound) {
		return "", domain.ErrUnauthorized
	}
	if err != nil {
		return "", fmt.Errorf("Login: %w", err)
	}

	if !ua.IsActive {
		return "", domain.ErrForbidden
	}
	if ua.AccountLockedUntil != nil && time.Now().Before(*ua.AccountLockedUntil) {
		return "", domain.ErrForbidden
	}

	// Быстрая проверка до открытия соединения (AD-11)
	if err := bcrypt.CompareHashAndPassword([]byte(ua.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrUnauthorized
	}

	conn, err := pgx.Connect(ctx, s.cfg.UserDSN(ua.Username, password))
	if err != nil {
		// pgx.Connect может вернуть ошибку при неверном пароле на стороне БД
		return "", domain.ErrUnauthorized
	}

	token, err := appjwt.Generate(ua.UserID, s.cfg.JWTSecret, s.cfg.SessionTTL)
	if err != nil {
		conn.Close(ctx)
		return "", fmt.Errorf("Login: generate token: %w", err)
	}

	s.store.Set(&session.Session{
		UserID:    ua.UserID,
		Username:  ua.Username,
		DBRole:    ua.RoleCode,
		Conn:      conn,
		ExpiresAt: time.Now().Add(s.cfg.SessionTTL),
	})

	logger.FromCtx(ctx).Info("login", "user_id", ua.UserID, "username", ua.Username, "role", ua.RoleCode)
	return token, nil
}

// Logout закрывает сессию пользователя.
func (s *AuthService) Logout(ctx context.Context, userID int) {
	s.store.Delete(userID)
	logger.FromCtx(ctx).Info("logout", "user_id", userID)
}
