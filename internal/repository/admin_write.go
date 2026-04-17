package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

// AdminInsertUserRole записывает назначение роли в v_user_roles_manage и возвращает
// username и db_role_name для последующего GRANT через RoleRepository (AD-9).
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

// AdminDeleteUserRole удаляет назначение роли и возвращает username и db_role_name
// для последующего REVOKE через RoleRepository (AD-9).
func (r *AppRepository) AdminDeleteUserRole(ctx context.Context, userID, roleID int) (username, dbRoleName string, err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", "", fmt.Errorf("AdminDeleteUserRole: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		SELECT u.username, r.db_role_name
		FROM users u, roles r
		WHERE u.user_id = $1 AND r.role_id = $2
	`, userID, roleID).Scan(&username, &dbRoleName)
	if err != nil {
		return "", "", fmt.Errorf("AdminDeleteUserRole: lookup: %w", err)
	}

	_, err = tx.Exec(ctx,
		`DELETE FROM v_user_roles_manage WHERE user_id = $1 AND role_id = $2`,
		userID, roleID,
	)
	if err != nil {
		return "", "", fmt.Errorf("AdminDeleteUserRole: delete: %w", err)
	}

	return username, dbRoleName, tx.Commit(ctx)
}

// AdminCreateUser создаёт запись пользователя в v_users_manage и возвращает user_id.
// Создание PostgreSQL-роли и установка пароля — отдельно через RoleRepository (AD-9, AD-10).
func (r *AppRepository) AdminCreateUser(ctx context.Context, dto domain.CreateUserDTO) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO v_users_manage (username, email, full_name, is_active)
		 VALUES ($1, $2, $3, TRUE) RETURNING user_id`,
		dto.Username, dto.Email, dto.FullName,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("AdminCreateUser: %w", err)
	}
	return id, nil
}

func (r *AppRepository) AdminSetUserActive(ctx context.Context, userID int, isActive bool) error {
	_, err := r.db.Exec(ctx, `UPDATE v_users_manage SET is_active=$1 WHERE user_id=$2`, isActive, userID)
	if err != nil {
		return fmt.Errorf("AdminSetUserActive: %w", err)
	}
	return nil
}

// AdminLockUntil устанавливает срок блокировки учётной записи.
// nil снимает блокировку.
func (r *AppRepository) AdminLockUntil(ctx context.Context, userID int, ts *time.Time) error {
	_, err := r.db.Exec(ctx, `UPDATE v_users_manage SET account_locked_until=$1 WHERE user_id=$2`, ts, userID)
	if err != nil {
		return fmt.Errorf("AdminLockUntil: %w", err)
	}
	return nil
}
