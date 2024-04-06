package config

import (
	"io"
	"log"
	"log/slog"
	"os"
)

func Logger(config Config) {
	path, _ := os.Getwd()

	slog.Info(path)

	replacer := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = source.File[len(path):]
		}
		return a
	}

	LOG_FILE := config.Log.Path
	// open log file
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		slog.Error("Unable to open log file.", err)
		log.Panic(err)
	}
	//defer logFile.Close()

	slogOptions := &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replacer,
	}

	textHandler := slog.NewTextHandler(
		io.MultiWriter(os.Stdout, logFile),
		slogOptions,
	).WithAttrs([]slog.Attr{slog.String("service", config.Server.Service)})

	logger := slog.New(textHandler)

	slog.SetDefault(logger)
}
