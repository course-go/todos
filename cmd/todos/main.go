package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/logger"
	"github.com/course-go/todos/internal/todos"
)

func main() {
	configPath := flag.String("config", "/etc/course-go/todos/config.yaml", "path to config file")
	flag.Parse()

	config, err := config.Parse(*configPath)
	if err != nil {
		slog.Error("failed parsing config",
			"error", err,
		)
		os.Exit(1)
	}

	logger, err := logger.New(config)
	if err != nil {
		slog.Error("failed creating logger",
			"error", err,
		)
		os.Exit(1)
	}

	ctx := context.Background()
	repository, err := todos.NewRepository(ctx, logger, config)
	if err != nil {
		logger.Error("failed creating todo repository",
			"error", err,
		)
		os.Exit(1)
	}

	mux := todos.NewRouter(logger, config, repository)
	hostname := fmt.Sprintf("%s:%s",
		config.Service.Host,
		config.Service.Port,
	)
	server := &http.Server{
		Addr:              hostname,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           mux,
	}
	logger.Info("starting server",
		"service", config.Service.Name,
		"hostname", hostname,
	)
	err = server.ListenAndServe()
	if err != nil {
		logger.Error("failed running server",
			"error", err,
		)
		os.Exit(1)
	}
}
