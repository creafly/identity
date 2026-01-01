package database

import (
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

type Migrator struct {
	db             *sqlx.DB
	migrationsPath string
}

func NewMigrator(db *sqlx.DB, migrationsPath string) *Migrator {
	return &Migrator{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

func (m *Migrator) Up() error {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	migration, err := migrate.NewWithDatabaseInstance(
		"file://"+m.migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Println("Migrations applied successfully")
	return nil
}

func (m *Migrator) Down() error {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	migration, err := migrate.NewWithDatabaseInstance(
		"file://"+m.migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := migration.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	log.Println("Migrations rolled back successfully")
	return nil
}

func (m *Migrator) Version() (uint, bool, error) {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return 0, false, err
	}

	migration, err := migrate.NewWithDatabaseInstance(
		"file://"+m.migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, err
	}

	return migration.Version()
}

func (m *Migrator) Steps(n int) error {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	migration, err := migrate.NewWithDatabaseInstance(
		"file://"+m.migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	return migration.Steps(n)
}

func (m *Migrator) Force(version int) error {
	driver, err := postgres.WithInstance(m.db.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	migration, err := migrate.NewWithDatabaseInstance(
		"file://"+m.migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	return migration.Force(version)
}
