package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoleRepository выполняет DDL-операции от имени суперпользователя (AD-9, AD-10).
// Используется исключительно для: CREATE/DROP ROLE, GRANT/REVOKE, ALTER ROLE PASSWORD.
// Синглтон на весь процесс — не использовать для бизнес-логики.
type RoleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

// CreateRole создаёт PostgreSQL login-роль с заданным паролем (AD-10).
func (s *RoleRepository) CreateRole(ctx context.Context, username, password string) error {
	sql := fmt.Sprintf(
		"CREATE ROLE %s LOGIN PASSWORD '%s'",
		pgx.Identifier{username}.Sanitize(),
		escapePgLiteral(password),
	)
	if _, err := s.pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("SystemRepo.CreateRole: %w", err)
	}
	return nil
}

// DropRole удаляет PostgreSQL login-роль (AD-10).
func (s *RoleRepository) DropRole(ctx context.Context, username string) error {
	sql := fmt.Sprintf("DROP ROLE IF EXISTS %s", pgx.Identifier{username}.Sanitize())
	if _, err := s.pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("SystemRepo.DropRole: %w", err)
	}
	return nil
}

// GrantGroupRole выдаёт групповую DB-роль пользователю (AD-10).
func (s *RoleRepository) GrantGroupRole(ctx context.Context, username, dbRoleName string) error {
	sql := fmt.Sprintf(
		"GRANT %s TO %s",
		pgx.Identifier{dbRoleName}.Sanitize(),
		pgx.Identifier{username}.Sanitize(),
	)
	if _, err := s.pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("SystemRepo.GrantGroupRole: %w", err)
	}
	return nil
}

// RevokeGroupRole отзывает групповую DB-роль у пользователя.
func (s *RoleRepository) RevokeGroupRole(ctx context.Context, username, dbRoleName string) error {
	sql := fmt.Sprintf(
		"REVOKE %s FROM %s",
		pgx.Identifier{dbRoleName}.Sanitize(),
		pgx.Identifier{username}.Sanitize(),
	)
	if _, err := s.pool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("SystemRepo.RevokeGroupRole: %w", err)
	}
	return nil
}

// SetPasswordAndHash атомарно обновляет bcrypt-хеш в таблице users
// и пароль PostgreSQL-роли через ALTER ROLE в одной транзакции (AD-10, AD-11).
func (s *RoleRepository) SetPasswordAndHash(ctx context.Context, userID int, username, bcryptHash, plainPassword string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("SystemRepo.SetPasswordAndHash: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx,
		"UPDATE users SET password_hash=$1, password_changed_at=NOW(), updated_at=NOW() WHERE user_id=$2",
		bcryptHash, userID,
	); err != nil {
		return fmt.Errorf("SystemRepo.SetPasswordAndHash: update hash: %w", err)
	}

	alterSQL := fmt.Sprintf(
		"ALTER ROLE %s PASSWORD '%s'",
		pgx.Identifier{username}.Sanitize(),
		escapePgLiteral(plainPassword),
	)
	if _, err = tx.Exec(ctx, alterSQL); err != nil {
		return fmt.Errorf("SystemRepo.SetPasswordAndHash: alter role: %w", err)
	}

	return tx.Commit(ctx)
}

// escapePgLiteral экранирует одинарные кавычки в строковых литералах PostgreSQL.
// Применяется только для DDL-параметров, которые не поддерживают $N-placeholders (пароли).
func escapePgLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
