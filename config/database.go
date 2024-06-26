package config

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
	"log/slog"
)

func Database() *sql.DB {
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		Config().Database.Username,
		Config().Database.Password,
		Config().Database.Hostname,
		Config().Database.Port,
		Config().Database.Database,
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		slog.Error("Failed to connect to database", err)
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		slog.Error("Failed to ping database", err)
		log.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Error("Failed to instantiate postgres driver", err)
		log.Fatal(err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		Config().Database.Migrations,
		"postgres", driver)
	if err != nil {
		slog.Error("database migration failed to create", err)
		log.Fatal(err)
	}

	if err = migration.Up(); err != nil {
		slog.Info(fmt.Sprintf("%s", err))
	}

	return db
}
