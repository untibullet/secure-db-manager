package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	applogger "github.com/untibullet/secure-db-manager/internal/logger"
	"github.com/untibullet/secure-db-manager/internal/seclog"
)

// RequestLogger обогащает контекст каждого запроса логгером с полями request_id, method, path (AD-14).
func RequestLogger(base *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			reqID := req.Header.Get("X-Request-Id")
			if reqID == "" {
				reqID = randomID()
			}

			l := base.With(
				"request_id", reqID,
				"method", req.Method,
				"path", req.URL.Path,
			)
			ctx := applogger.WithLogger(req.Context(), l)
			ctx = seclog.WithClientIP(ctx, c.RealIP())
			c.SetRequest(req.WithContext(ctx))

			start := time.Now()
			err := next(c)
			elapsed := time.Since(start)

			status := c.Response().Status
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				}
			}

			l.Info("request",
				"status", status,
				"duration_ms", elapsed.Milliseconds(),
			)
			return err
		}
	}
}

func randomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
