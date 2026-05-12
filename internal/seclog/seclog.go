package seclog

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type SecurityEvent struct {
	Timestamp time.Time `json:"timestamp"`
	EventType string    `json:"event_type"`
	Username  string    `json:"username,omitempty"`
	UserID    int       `json:"user_id,omitempty"`
	ClientIP  string    `json:"client_ip,omitempty"`
	DBRole    string    `json:"db_role,omitempty"`
	Reason    string    `json:"reason,omitempty"`
	Path      string    `json:"path,omitempty"`
}

type SecurityLogger struct {
	mu  sync.Mutex
	enc *json.Encoder
	f   *os.File
}

func New(path string) (*SecurityLogger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &SecurityLogger{enc: json.NewEncoder(f), f: f}, nil
}

// Noop returns a SecurityLogger that discards all events. Use in tests.
func Noop() *SecurityLogger {
	f, _ := os.Open(os.DevNull)
	return &SecurityLogger{enc: json.NewEncoder(f), f: f}
}

func (sl *SecurityLogger) Write(e SecurityEvent) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	sl.mu.Lock()
	defer sl.mu.Unlock()
	_ = sl.enc.Encode(e)
}

func (sl *SecurityLogger) Close() error {
	return sl.f.Close()
}

// ── Context helpers ────────────────────────────────────────────────────────────

type clientIPKey struct{}

func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

func ClientIPFromCtx(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}
