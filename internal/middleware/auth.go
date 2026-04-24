package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	appjwt "github.com/untibullet/secure-db-manager/internal/jwt"
	"github.com/untibullet/secure-db-manager/internal/logger"
	"github.com/untibullet/secure-db-manager/internal/session"
)

type sessionKey struct{}

// Auth валидирует JWT из заголовка Authorization, извлекает сессию из store
// и кладёт *pgx.Conn в контекст запроса (AD-8).
func Auth(secret string, store *session.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			claims, err := appjwt.Parse(strings.TrimPrefix(header, "Bearer "), secret)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			userID, err := claims.UserID()
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			sess, ok := store.Get(userID)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, sessionKey{}, sess)
			ctx = logger.WithLogger(ctx, logger.FromCtx(ctx).With(
				"user_id", sess.UserID,
				"db_role", sess.DBRole,
			))
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// SessionFromCtx извлекает сессию из контекста запроса.
func SessionFromCtx(ctx context.Context) *session.Session {
	sess, _ := ctx.Value(sessionKey{}).(*session.Session)
	return sess
}

// ConnFromCtx извлекает *pgx.Conn из сессии в контексте запроса.
func ConnFromCtx(ctx context.Context) *pgx.Conn {
	if sess := SessionFromCtx(ctx); sess != nil {
		return sess.Conn
	}
	return nil
}
