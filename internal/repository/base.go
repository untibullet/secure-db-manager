package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/untibullet/secure-db-manager/internal/domain"
)

// mapRepoError преобразует PG-ошибку 42501 (insufficient_privilege) в domain.ErrForbidden.
// Применяется в write-методах, работающих через view с привилегиями по роли.
func mapRepoError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42501" {
		return domain.ErrForbidden
	}
	return err
}

// safeColRe допускает только snake_case идентификаторы.
// Ключи Filter должны быть хардкодированными константами — никогда не пользовательский ввод (AD-5).
var safeColRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type AppRepository struct {
	db DBTX
}

func NewAppRepository(db DBTX) *AppRepository {
	return &AppRepository{db: db}
}

func buildWhereClause(filter domain.Filter, paging *domain.Paging) (string, []any, error) {
	var conditions []string
	var args []any
	idx := 1

	keys := make([]string, 0, len(filter))
	for k := range filter {
		if !safeColRe.MatchString(k) {
			return "", nil, fmt.Errorf("buildWhereClause: небезопасное имя колонки %q — ключи Filter должны быть хардкодированными константами", k)
		}
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		conditions = append(conditions, fmt.Sprintf("%s = $%d", k, idx))
		args = append(args, filter[k])
		idx++
	}

	query := ""
	if len(conditions) > 0 {
		query = " WHERE " + strings.Join(conditions, " AND ")
	}

	if paging != nil {
		if paging.Limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d", idx)
			args = append(args, paging.Limit)
			idx++
		}
		if paging.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", idx)
			args = append(args, paging.Offset)
		}
	}

	return query, args, nil
}
