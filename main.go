package main

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go_auth/config"
	"go_auth/user"
	"log"
	"net/http"
	"os"
)

func main() {
	cfg := config.ReadConfig()

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
		log.Println(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})

	if err != nil {
		log.Println(err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		cfg.Database.Migrations,
		"postgres", driver)

	if err != nil {
		log.Println(err)
	}

	err = migration.Up()

	if err != nil {
		log.Println(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/", func(writer http.ResponseWriter, request *http.Request) {
		usr := user.Create("derek", "password")
		insert, err := user.Insert(db, usr)
		if err != nil {
			log.Println(err)
		}
		response := fmt.Sprintf("Welcome %s", insert.Username)
		writer.Write([]byte(response))
	})

	log.Printf("Listening on port %s", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), router)

	if err != nil {
		log.Println(err)
	}
}
