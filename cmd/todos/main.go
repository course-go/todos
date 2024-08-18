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
	"github.com/course-go/todos/internal/controllers"
	"github.com/course-go/todos/internal/logger"
	"github.com/course-go/todos/internal/repository"
	ttime "github.com/course-go/todos/internal/time"
)

var Version string

var (
	versionFlag    = flag.Bool("version", false, "output program version and exit")
	configPathFlag = flag.String("config", "/etc/course-go/todos/config.yaml", "path to config file")
)

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Printf("TODOS: [%s]\n", Version) //nolint: forbidigo
		os.Exit(0)
	}

	config, err := config.Parse(*configPathFlag)
	if err != nil {
		slog.Error("failed parsing config",
			"error", err,
		)
		os.Exit(1)
	}

	logger, err := logger.New(&config.Logging)
	if err != nil {
		slog.Error("failed creating logger",
			"error", err,
		)
		os.Exit(1)
	}

	err = repository.Migrate(&config.Database, logger)
	if err != nil {
		logger.Error("failed migrating database",
			"error", err,
		)
		os.Exit(1)
	}

	ctx := context.Background()
	repo, err := repository.New(ctx, logger, &config.Database)
	if err != nil {
		logger.Error("failed creating todo repository",
			"error", err,
		)
		os.Exit(1)
	}

	mux := controllers.NewAPIRouter(logger, config, ttime.Now(), repo)
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
		"version", Version,
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
