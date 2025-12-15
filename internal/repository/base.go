package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Paging определяет параметры пагинации
type Paging struct {
	Limit  int
	Offset int
}

// Filter - карта фильтров (ключ: колонка, значение: значение)
type Filter map[string]interface{}

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Repository - основная структура слоя данных
type Repository struct {
	db DBTX // Интерфейс, скрывающий *pgx.Conn, *pgxpool.Pool или pgx.Tx
}

func NewRepository(db DBTX) *Repository {
	return &Repository{db: db}
}

// Вспомогательная функция для построения WHERE и аргументов
func buildWhereClause(filter Filter, paging *Paging) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	idx := 1

	for k, v := range filter {
		// Простая реализация равенства. Для LIKE, >, < нужна более сложная логика
		conditions = append(conditions, fmt.Sprintf("%s = $%d", k, idx))
		args = append(args, v)
		idx++
	}

	query := ""
	if len(conditions) > 0 {
		query = " WHERE " + strings.Join(conditions, " AND ")
	} else {
		query = " WHERE 1=1 " // Чтобы всегда можно было добавить LIMIT/OFFSET
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

	return query, args
}
