package session

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// Session хранит активное подключение пользователя (AD-8).
// Ключ хранилища — UserID, одна сессия на пользователя.
type Session struct {
	UserID    int
	Username  string
	DBRole    string
	Conn      *pgx.Conn
	ExpiresAt time.Time
}

// Store — in-memory хранилище сессий (AD-8).
// Не масштабируется горизонтально; при перезапуске все сессии инвалидируются.
type Store struct {
	mu       sync.RWMutex
	sessions map[int]*Session
}

func NewStore() *Store {
	return &Store{sessions: make(map[int]*Session)}
}

func (s *Store) Get(userID int) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[userID]
	if !ok || time.Now().After(sess.ExpiresAt) {
		return nil, false
	}
	return sess, true
}

// Set сохраняет сессию, закрывая предыдущую при наличии (AD-8).
func (s *Store) Set(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.sessions[sess.UserID]; ok {
		old.Conn.Close(context.Background())
	}
	s.sessions[sess.UserID] = sess
}

func (s *Store) Delete(userID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[userID]; ok {
		sess.Conn.Close(context.Background())
		delete(s.sessions, userID)
	}
}

// StartReaper запускает фоновую горутину, удаляющую протухшие сессии каждые interval.
func (s *Store) StartReaper(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.reap()
			}
		}
	}()
}

func (s *Store) reap() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, sess := range s.sessions {
		if now.After(sess.ExpiresAt) {
			sess.Conn.Close(context.Background())
			delete(s.sessions, id)
		}
	}
}
