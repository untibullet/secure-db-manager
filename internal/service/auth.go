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
	"github.com/untibullet/secure-db-manager/internal/seclog"
	"github.com/untibullet/secure-db-manager/internal/session"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepo interface {
	GetUserForLogin(ctx context.Context, username string) (*domain.UserAuth, error)
}

type AuthService struct {
	store  *session.Store
	cfg    *config.Config
	seclog *seclog.SecurityLogger
}

func NewAuthService(store *session.Store, cfg *config.Config, sl *seclog.SecurityLogger) *AuthService {
	return &AuthService{store: store, cfg: cfg, seclog: sl}
}

// Login проверяет пароль через bcrypt, открывает pgx.Conn от имени пользователя,
// сохраняет сессию и возвращает JWT (AD-8, AD-11).
func (s *AuthService) Login(ctx context.Context, repo AuthRepo, username, password string) (string, error) {
	ip := seclog.ClientIPFromCtx(ctx)

	ua, err := repo.GetUserForLogin(ctx, username)
	if errors.Is(err, domain.ErrNotFound) {
		s.seclog.Write(seclog.SecurityEvent{EventType: "auth.login.failed", Username: username, ClientIP: ip, Reason: "user_not_found"})
		return "", domain.ErrUnauthorized
	}
	if err != nil {
		return "", fmt.Errorf("Login: %w", err)
	}

	if !ua.IsActive {
		s.seclog.Write(seclog.SecurityEvent{EventType: "auth.login.failed", Username: ua.Username, UserID: ua.UserID, ClientIP: ip, Reason: "inactive"})
		return "", domain.ErrForbidden
	}
	if ua.AccountLockedUntil != nil && time.Now().Before(*ua.AccountLockedUntil) {
		s.seclog.Write(seclog.SecurityEvent{EventType: "auth.login.failed", Username: ua.Username, UserID: ua.UserID, ClientIP: ip, Reason: "locked"})
		return "", domain.ErrForbidden
	}

	// Быстрая проверка до открытия соединения (AD-11)
	if err := bcrypt.CompareHashAndPassword([]byte(ua.PasswordHash), []byte(password)); err != nil {
		s.seclog.Write(seclog.SecurityEvent{EventType: "auth.login.failed", Username: ua.Username, UserID: ua.UserID, ClientIP: ip, Reason: "bcrypt"})
		return "", domain.ErrUnauthorized
	}

	conn, err := pgx.Connect(ctx, s.cfg.UserDSN(ua.Username, password))
	if err != nil {
		// pgx.Connect может вернуть ошибку при неверном пароле на стороне БД
		s.seclog.Write(seclog.SecurityEvent{EventType: "auth.login.failed", Username: ua.Username, UserID: ua.UserID, ClientIP: ip, Reason: "pg_auth"})
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

	s.seclog.Write(seclog.SecurityEvent{EventType: "auth.login.success", Username: ua.Username, UserID: ua.UserID, ClientIP: ip, DBRole: ua.RoleCode})
	logger.FromCtx(ctx).Info("login", "user_id", ua.UserID, "username", ua.Username, "role", ua.RoleCode)
	return token, nil
}

// Logout закрывает сессию пользователя.
func (s *AuthService) Logout(ctx context.Context, userID int, username string) {
	s.store.Delete(userID)
	s.seclog.Write(seclog.SecurityEvent{EventType: "auth.logout", UserID: userID, Username: username, ClientIP: seclog.ClientIPFromCtx(ctx)})
	logger.FromCtx(ctx).Info("logout", "user_id", userID)
}
