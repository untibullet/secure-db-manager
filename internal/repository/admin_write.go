package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) AdminGrantRole(ctx context.Context, userID, roleID int, validUntil time.Time) error {
	// r.db должен быть *pgxpool.Pool или *pgx.Conn, поддерживающим Begin
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Получаем системные имена пользователя и роли для команды GRANT
	// Нам нужны именно имена в БД (например, 'ivan_ivanov', 'db_tester')
	var userName, roleName string
	queryNames := `
        SELECT u.username, r.code 
        FROM users u, roles r 
        WHERE u.user_id = $1 AND r.role_id = $2
    `
	err = tx.QueryRow(ctx, queryNames, userID, roleID).Scan(&userName, &roleName)
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
	// ВНИМАНИЕ: Ident (Identifier) используется для безопасной подстановки имен объектов
	grantSql := fmt.Sprintf("GRANT %s TO %s", pgx.Identifier{roleName}.Sanitize(), pgx.Identifier{userName}.Sanitize())

	_, err = tx.Exec(ctx, grantSql)
	if err != nil {
		return fmt.Errorf("failed to execute system grant: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *Repository) AdminSetUserActive(ctx context.Context, userID int, isActive bool) error {
	// Обновляем через главное админское представление
	sql := `UPDATE v_admin_users_and_roles SET is_active=$1 WHERE user_id=$2`
	_, err := r.db.Exec(ctx, sql, isActive, userID)
	return err
}

// TODO: удалить этот метод или добавить поле lock_until
func (r *Repository) AdminLockUntil(ctx context.Context, userID int, ts time.Time) error {
	// В v_admin_users_and_roles нет поля lock_until явно в SELECT, но если триггер поддерживает UPDATE,
	// или если view содержит это поле (в файле не показано, но подразумевается логикой блокировки):
	sql := `UPDATE v_admin_users_and_roles SET locked_until=$1 WHERE user_id=$2`
	_, err := r.db.Exec(ctx, sql, ts, userID)
	return err
}
