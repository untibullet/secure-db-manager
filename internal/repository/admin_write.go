package repository

import (
	"context"
	"fmt"
	"time"
)

// AdminInsertUserRole записывает назначение роли в v_user_roles_manage и возвращает
// username и db_role_name для последующего GRANT через SystemRepo (AD-9).
// GRANT выполняется отдельно сервисным слоем через SystemRepo.GrantGroupRole.
func (r *AppRepository) AdminInsertUserRole(ctx context.Context, userID, roleID int, validUntil *time.Time) (username, dbRoleName string, err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", "", fmt.Errorf("AdminInsertUserRole: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		SELECT u.username, r.db_role_name
		FROM users u, roles r
		WHERE u.user_id = $1 AND r.role_id = $2
	`, userID, roleID).Scan(&username, &dbRoleName)
	if err != nil {
		return "", "", fmt.Errorf("AdminInsertUserRole: lookup: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO v_user_roles_manage (user_id, role_id, valid_until) VALUES ($1, $2, $3)`,
		userID, roleID, validUntil,
	)
	if err != nil {
		return "", "", fmt.Errorf("AdminInsertUserRole: insert: %w", err)
	}

	return username, dbRoleName, tx.Commit(ctx)
}

func (r *AppRepository) AdminSetUserActive(ctx context.Context, userID int, isActive bool) error {
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
func (r *AppRepository) AdminLockUntil(ctx context.Context, userID int, ts *time.Time) error {
	sql := `UPDATE v_users_manage SET account_locked_until=$1 WHERE user_id=$2`
	_, err := r.db.Exec(ctx, sql, ts, userID)
	if err != nil {
		return fmt.Errorf("AdminLockUntil: %w", err)
	}
	return nil
}
