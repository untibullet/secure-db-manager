package repository

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Paging определяет параметры пагинации
type Paging struct {
	Limit  int
	Offset int
}

// Filter - карта фильтров (ключ: колонка, значение: значение)
type Filter map[string]interface{}

// Repository - основная структура слоя данных
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
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
