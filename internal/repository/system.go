package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/untibullet/secure-db-manager/internal/domain"
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

// ListAvailableRoles возвращает все бизнес-роли из таблицы roles через суперпользовательский пул.
func (s *RoleRepository) ListAvailableRoles(ctx context.Context) ([]domain.Role, error) {
	rows, err := s.pool.Query(ctx, `SELECT role_id, code, name FROM roles ORDER BY role_id`)
	if err != nil {
		return nil, fmt.Errorf("ListAvailableRoles: %w", err)
	}
	defer rows.Close()
	var roles []domain.Role
	for rows.Next() {
		var r domain.Role
		if err := rows.Scan(&r.ID, &r.Code, &r.Name); err != nil {
			return nil, fmt.Errorf("ListAvailableRoles: %w", err)
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

// GetUserForLogin возвращает данные, необходимые для аутентификации (AD-11).
// Выполняется через суперпользовательский пул — единственный способ прочитать password_hash.
func (s *RoleRepository) GetUserForLogin(ctx context.Context, username string) (*domain.UserAuth, error) {
	sql := `
		SELECT u.user_id, u.username, u.password_hash, u.is_active, u.account_locked_until,
		       COALESCE(r.code, '') AS role_code
		FROM users u
		LEFT JOIN user_roles ur ON u.user_id = ur.user_id
		    AND (ur.valid_until IS NULL OR ur.valid_until > NOW())
		LEFT JOIN roles r ON ur.role_id = r.role_id
		WHERE u.username = $1
		ORDER BY ur.granted_at DESC NULLS LAST
		LIMIT 1
	`
	var ua domain.UserAuth
	err := s.pool.QueryRow(ctx, sql, username).Scan(
		&ua.UserID, &ua.Username, &ua.PasswordHash,
		&ua.IsActive, &ua.AccountLockedUntil, &ua.RoleCode,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetUserForLogin: %w", err)
	}
	return &ua, nil
}

// AdminCreateUser атомарно создаёт запись в таблице users и PostgreSQL login-роль (AD-10).
// Не передавать bcryptHash и plainPassword в логи.
func (s *RoleRepository) AdminCreateUser(ctx context.Context, dto domain.CreateUserDTO, bcryptHash string) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("AdminCreateUser: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID int
	err = tx.QueryRow(ctx,
		`INSERT INTO users (username, email, full_name, is_active, password_hash)
		 VALUES ($1, $2, $3, TRUE, $4) RETURNING user_id`,
		dto.Username, dto.Email, dto.FullName, bcryptHash,
	).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("AdminCreateUser: insert user: %w", err)
	}

	createSQL := fmt.Sprintf(
		"CREATE ROLE %s LOGIN PASSWORD '%s'",
		pgx.Identifier{dto.Username}.Sanitize(),
		escapePgLiteral(dto.Password),
	)
	if _, err = tx.Exec(ctx, createSQL); err != nil {
		return 0, fmt.Errorf("AdminCreateUser: create role: %w", err)
	}

	return userID, tx.Commit(ctx)
}

// escapePgLiteral экранирует одинарные кавычки в строковых литералах PostgreSQL.
// Применяется только для DDL-параметров, которые не поддерживают $N-placeholders (пароли).
func escapePgLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
