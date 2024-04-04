package config

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
)

func Database(cfg Config) *sql.DB {
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Hostname,
		cfg.Database.Port,
		cfg.Database.Database,
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		cfg.Database.Migrations,
		"postgres", driver)
	if err != nil {
		log.Println(err)
	}

	if err = migration.Up(); err != nil {
		log.Println(err)
	}

	return db
}
