package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (r *AppRepository) AdminListUsersAndRoles(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.UserAdminView, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("AdminListUsersAndRoles: %w", err)
	}
	sql := `SELECT user_id, username, full_name, email, is_active, last_login, roles
	        FROM v_admin_users_and_roles` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("AdminListUsersAndRoles: %w", err)
	}
	defer rows.Close()

	var users []domain.UserAdminView
	for rows.Next() {
		var u domain.UserAdminView
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Email, &u.IsActive, &u.LastLogin, &u.Roles); err != nil {
			return nil, fmt.Errorf("AdminListUsersAndRoles: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *AppRepository) AdminGetUserByID(ctx context.Context, id int) (*domain.UserAdminView, error) {
	sql := `SELECT user_id, username, full_name, email, is_active, last_login, roles
	        FROM v_admin_users_and_roles WHERE user_id = $1`
	var u domain.UserAdminView
	err := r.db.QueryRow(ctx, sql, id).Scan(&u.ID, &u.Username, &u.FullName, &u.Email, &u.IsActive, &u.LastLogin, &u.Roles)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("AdminGetUserByID: %w", err)
	}
	return &u, nil
}

func (r *AppRepository) AdminListAuditLog(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.AuditEntry, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("AdminListAuditLog: %w", err)
	}
	sql := `SELECT audit_id, table_name, operation, record_id, user_id, changed_at, old_values, new_values, ip_address
	        FROM v_admin_audit_log` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("AdminListAuditLog: %w", err)
	}
	defer rows.Close()

	var entries []domain.AuditEntry
	for rows.Next() {
		var e domain.AuditEntry
		if err := rows.Scan(&e.AuditID, &e.TableName, &e.Operation, &e.RecordID, &e.UserID, &e.ChangedAt, &e.OldValues, &e.NewValues, &e.IPAddress); err != nil {
			return nil, fmt.Errorf("AdminListAuditLog: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
