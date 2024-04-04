package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	db := config.Database(cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	config.Prometheus(router)
	user.Routes(cfg, router, db)

	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/*", http.StripPrefix("/static/", fs))

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		log.Println(err)
	}
}
