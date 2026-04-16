package repository

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// safeColRe допускает только snake_case идентификаторы.
// Filter-ключи должны быть хардкодированными константами в репозитории —
// никогда не передавать пользовательский ввод напрямую в Filter.
var safeColRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Paging определяет параметры пагинации
type Paging struct {
	Limit  int
	Offset int
}

// Filter - карта фильтров (ключ: колонка, значение: значение)
type Filter map[string]any

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Repository - основная структура слоя данных
type Repository struct {
	db DBTX // Интерфейс, скрывающий *pgx.Conn, *pgxpool.Pool или pgx.Tx
}

func NewRepository(db DBTX) *Repository {
	return &Repository{db: db}
}

// buildWhereClause строит WHERE-клаузу из фильтра.
// ВАЖНО: ключи Filter должны быть хардкодированными строковыми литералами.
// Передача пользовательского ввода в качестве ключа — программная ошибка; функция вернёт error.
func buildWhereClause(filter Filter, paging *Paging) (string, []any, error) {
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
			idx++
		}
	}

	return query, args, nil
}
