package services

import (
	"context"
	"fmt"
	"github.com/fsdevblog/gophkeeper/internal/config"
	"github.com/fsdevblog/gophkeeper/internal/db/migrations"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcauth"
	"github.com/fsdevblog/gophkeeper/internal/storage/services/svcentry"
	"github.com/fsdevblog/gophkeeper/internal/storage/uow"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Collection struct {
	config       *config.Config
	AuthService  *svcauth.AuthService
	EntryService *svcentry.EntryService
}

func NewCollection(ctx context.Context, config *config.Config) (*Collection, error) {
	c := &Collection{
		config: config,
	}
	conn, errConn := c.dbUp(ctx)
	if errConn != nil {
		return nil, fmt.Errorf("failed to dbUp: %w", errConn)
	}
	c.initServices(conn)
	return c, nil
}

func (c *Collection) initServices(conn *pgxpool.Pool) {
	unitOfWork := uow.New(conn)
	c.AuthService = svcauth.New(unitOfWork, []byte(c.config.JWTSecret), func(opt *svcauth.Options) {
		opt.JWTTokenExpiration = time.Duration(c.config.JWTExpireInSeconds) * time.Second
	})
	c.EntryService = svcentry.New(unitOfWork)
}

func (c *Collection) dbUp(ctx context.Context) (*pgxpool.Pool, error) {
	pool, errPool := newPostgresConnection(ctx, c.config.DatabaseDSN)
	if errPool != nil {
		return nil, fmt.Errorf("failed to create postgres connection: %w", errPool)
	}

	if err := migrations.PostgresMigrate(c.config.DatabaseDSN); err != nil {
		return nil, fmt.Errorf("failed to migrate postgres schema: %w", err)
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
