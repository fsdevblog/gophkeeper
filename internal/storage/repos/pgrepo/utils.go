package pgrepo

import (
	"context"
	"fmt"
	"github.com/fsdevblog/gophkeeper/internal/db/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func DBUp(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, errPool := newPostgresConnection(ctx, dsn)
	if errPool != nil {
		return nil, fmt.Errorf("postgres connection: %w", errPool)
	}

	if err := migrations.PostgresMigrate(dsn); err != nil {
		return nil, fmt.Errorf("migrate postgres schema: %w", err)
	}
	return pool, nil
}

// newPostgresConnection создает новый пул подключений к PostgreSQL.
//
// Параметры:
//   - ctx: контекст выполнения
//   - dsn: строка подключения к базе данных (Data Source Name)
//
// Возвращает:
//   - *pgxpool.Pool: пул подключений к PostgreSQL
//   - error: ошибка создания подключения
func newPostgresConnection(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolConfig, confErr := pgxpool.ParseConfig(dsn)
	if confErr != nil {
		return nil, fmt.Errorf("failed to parse config: %w", confErr)
	}
	pool, poolErr := pgxpool.NewWithConfig(ctx, poolConfig)
	if poolErr != nil {
		return nil, fmt.Errorf("failed to create pool: %w", poolErr)
	}
	return pool, nil
}
