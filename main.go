package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go_auth/config"
	"go_auth/user"
	"log"
	"net/http"
	"os"
)

func main() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	cfg := config.ReadConfig()
	db := config.Database(cfg)
	config.Prometheus(cfg, router)
	user.Routes(cfg, router, db)

	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/*", http.StripPrefix("/static/", fs))

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		log.Println(err)
	}
}
