package config

import (
	"fmt"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"log"
	"log/slog"
	"os"
)

var loggerShutDown func()

func LoggerShutDown() {
	loggerShutDown()
}

func Logger() {
	path, _ := os.Getwd()

	slog.Info(path)

	LOG_FILE := Config().Log.Path
	// open log file
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		msg := fmt.Sprintf("Unable to open log file `%s`", LOG_FILE)
		slog.Error(msg, err)
		log.Panic(err)
	}
	loggerShutDown = func() {
		logFile.Close()
	}

	handler := otelslog.NewHandler("default-logger")

	logger := slog.New(handler)

	slog.SetDefault(logger)
}
