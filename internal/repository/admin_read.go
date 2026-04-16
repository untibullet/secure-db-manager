package repository

import (
	"context"
	"fmt"
)

func (r *Repository) AdminListUsersAndRoles(ctx context.Context, filter Filter, paging Paging) ([]UserAdminView, error) {
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

	var users []UserAdminView
	for rows.Next() {
		var u UserAdminView
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Email, &u.IsActive, &u.LastLogin, &u.Roles); err != nil {
			return nil, fmt.Errorf("AdminListUsersAndRoles: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
