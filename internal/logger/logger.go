package logger

import (
	"context"
	"log/slog"
)

type contextKey struct{}

// WithLogger кладёт логгер в контекст (AD-14).
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromCtx возвращает логгер из контекста. Fallback: slog.Default() — никогда не nil (AD-14).
func FromCtx(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(contextKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
