package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) AdminGrantRole(ctx context.Context, userID, roleID int, validUntil *time.Time) error {
	// r.db должен быть *pgxpool.Pool или *pgx.Conn, поддерживающим Begin
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Получаем username и db_role_name для команды GRANT
	var userName, dbRoleName string
	queryNames := `
        SELECT u.username, r.db_role_name
        FROM users u, roles r
        WHERE u.user_id = $1 AND r.role_id = $2
    `
	err = tx.QueryRow(ctx, queryNames, userID, roleID).Scan(&userName, &dbRoleName)
	if err != nil {
		return fmt.Errorf("failed to get names for grant: %w", err)
	}

	// Выполняем бизнес-логику: запись в таблицу учета ролей
	insertSql := `INSERT INTO v_user_roles_manage (user_id, role_id, valid_until) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, insertSql, userID, roleID, validUntil)
	if err != nil {
		return fmt.Errorf("failed to insert user role record: %w", err)
	}

	// Выполняем системную логику: GRANT
	// pgx.Identifier.Sanitize() защищает от SQL injection при подстановке идентификаторов
	grantSql := fmt.Sprintf("GRANT %s TO %s", pgx.Identifier{dbRoleName}.Sanitize(), pgx.Identifier{userName}.Sanitize())

	_, err = tx.Exec(ctx, grantSql)
	if err != nil {
		return fmt.Errorf("failed to execute system grant: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *Repository) AdminSetUserActive(ctx context.Context, userID int, isActive bool) error {
	sql := `UPDATE v_users_manage SET is_active=$1 WHERE user_id=$2`
	_, err := r.db.Exec(ctx, sql, isActive, userID)
	if err != nil {
		return fmt.Errorf("AdminSetUserActive: %w", err)
	}
	return nil
}

// AdminLockUntil устанавливает срок блокировки учётной записи.
// nil снимает блокировку, не-nil указатель устанавливает конкретный срок.
// Для бессрочной блокировки передайте указатель на pgtype.Timestamptz{InfinityModifier: pgtype.Infinity}.
func (r *Repository) AdminLockUntil(ctx context.Context, userID int, ts *time.Time) error {
	sql := `UPDATE v_users_manage SET account_locked_until=$1 WHERE user_id=$2`
	_, err := r.db.Exec(ctx, sql, ts, userID)
	if err != nil {
		return fmt.Errorf("AdminLockUntil: %w", err)
	}
	return nil
}
