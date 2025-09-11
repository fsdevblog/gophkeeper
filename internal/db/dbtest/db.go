package dbtest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fsdevblog/gophkeeper/internal/db/migrations"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // драйвер для database/sql
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PgConnect struct {
	PgContainer *postgres.PostgresContainer
	PgPool      *pgxpool.Pool
	Fixtures    *testfixtures.Loader
}

type ConnectionConfig struct {
	DatabaseName     string
	DatabaseUsername string
	DatabasePassword string
	DockerImage      string
	StartupTimeout   time.Duration
	FixturesPath     string // if empty, fixtures will not be loaded
}

const (
	defaultDBName         = "test_db"
	defaultDBUser         = "testuser"
	defaultDBPass         = "password"
	defaultImage          = "postgres:17"
	defaultStartupTimeout = 15 * time.Second
	defaultFixturesPath   = "testdata/fixtures"
)

// Connect поднимает постгрес докер контейнер, выполняет миграции и возвращает PgConnect и ошибку.
func Connect(ctx context.Context, opts ...func(*ConnectionConfig)) (*PgConnect, error) {
	config := ConnectionConfig{
		DatabaseName:     defaultDBName,
		DatabaseUsername: defaultDBUser,
		DatabasePassword: defaultDBPass,
		DockerImage:      defaultImage,
		StartupTimeout:   defaultStartupTimeout,
		FixturesPath:     defaultFixturesPath,
	}
	for _, opt := range opts {
		opt(&config)
	}
	pgContainer, errPgContainer := postgres.Run(ctx, config.DockerImage,
		postgres.WithDatabase(config.DatabaseName),
		postgres.WithUsername(config.DatabaseUsername),
		postgres.WithPassword(config.DatabasePassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("ready").
				WithOccurrence(2).WithStartupTimeout(config.StartupTimeout)),
	)
	if errPgContainer != nil {
		return nil, fmt.Errorf("failed to start postgres container: %s", errPgContainer.Error())
	}

	dsn, errDsn := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if errDsn != nil {
		return nil, fmt.Errorf("failed to get connection string: %s", errDsn.Error())
	}

	pgPool, errPool := connectPg(ctx, dsn)
	if errPool != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %s", errPool.Error())
	}

	// Создаем стандартное подключение sql.DB для testfixtures
	sqlDB, errSQLdb := sql.Open("pgx", dsn)
	if errSQLdb != nil {
		return nil, fmt.Errorf("failed to open sql.DB: %s", errSQLdb.Error())
	}

	errMigrations := migrations.PostgresMigrate(dsn)
	if errMigrations != nil {
		return nil, fmt.Errorf("failed to migrate postgres schema: %s", errMigrations.Error())
	}

	fixtures, errFixtures := createFixtures(sqlDB, config.FixturesPath)

	if errFixtures != nil {
		return nil, fmt.Errorf("failed to create testfixtures: %s", errFixtures.Error())
	}

	return &PgConnect{
		PgContainer: pgContainer,
		PgPool:      pgPool,
		Fixtures:    fixtures,
	}, nil
}

func ClearTables(conn *pgxpool.Pool, tables ...string) error {
	var errs error
	for _, table := range tables {
		_, err := conn.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("clear table %s: %w", table, err))
		}
	}
	return errs
}

func createFixtures(sqlDB *sql.DB, fixturesPath string) (*testfixtures.Loader, error) {
	options := []func(*testfixtures.Loader) error{
		testfixtures.Database(sqlDB),
		testfixtures.Dialect("postgres"),
	}

	if fixturesPath != "" {
		options = append(options, testfixtures.Directory(fixturesPath))
	}

	fixtures, err := testfixtures.New(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create testfixtures: %w", err)
	}

	return fixtures, nil
}

// connectPg создает подключение к бд, и пингует его 10 попыток с интервалом в 300 мс.
func connectPg(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolConfig, errConf := pgxpool.ParseConfig(dsn)
	if errConf != nil {
		return nil, fmt.Errorf("parse postgres config: %s", errConf.Error())
	}
	pool, errPool := pgxpool.NewWithConfig(ctx, poolConfig)
	if errPool != nil {
		return nil, fmt.Errorf("failed to create pool: %s", errPool.Error())
	}
	var errPing = make(chan error, 1)
	wg := new(sync.WaitGroup)
	wg.Add(1)

	maxAttempts := 10
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		attempts := 0
		for {
			err := pool.Ping(ctx)
			if err == nil {
				errPing <- nil
				return
			}
			time.Sleep(300 * time.Millisecond) //nolint:mnd
			attempts++
			if attempts > maxAttempts {
				errPing <- fmt.Errorf("failed to ping postgres: %s", err.Error())
				return
			}
		}
	}(wg)
	wg.Wait()

	close(errPing)
	if errPingErr := <-errPing; errPingErr != nil {
		return nil, errPingErr
	}

	return pool, nil
}
