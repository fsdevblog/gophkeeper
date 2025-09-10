package migrations

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // драйвер postgres для golang-migrate
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationsFS embed.FS

// PostgresMigrate выполняет миграцию вверх базы данных.
func PostgresMigrate(dsn string) error {
	d, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if errMigrate := m.Up(); errMigrate != nil && !errors.Is(errMigrate, migrate.ErrNoChange) {
		return fmt.Errorf("failed to migrate schema: %s", errMigrate.Error())
	}
	return nil
}
