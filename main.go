package main

import (
	"fmt"
	"go_auth/config"
	"go_auth/user"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	cfg := config.ReadConfig()
	config.Logger(cfg)
	db := config.Database(cfg)

	router := http.NewServeMux()
	handler := config.Prometheus(cfg, router)

	user.Routes(cfg, router, db)

	fileServer := http.FileServer(http.Dir("static"))
	router.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
	// 	writer.Write([]byte("/"))
	// })

	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}

	server := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	slog.Info(fmt.Sprintf("server listening at %s", port))
	if err := server.ListenAndServe(); err != nil {
		msg := fmt.Sprintf("error while serving: %s", err)
		slog.Error(msg)
		log.Panicf(msg)
	}

}
