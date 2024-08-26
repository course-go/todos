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
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
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

	location, err := time.LoadLocation(config.Service.Location)
	if err != nil {
		slog.Error("failed loading location",
			"error", err,
			"location", config.Service.Location,
		)
		os.Exit(1)
	}

	time.Local = location //nolint: gosmopolitan

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

	exporter, err := prometheus.New()
	if err != nil {
		logger.Error("failed creating prometheus exporter",
			"error", err,
		)
		os.Exit(1)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	mux, err := controllers.NewAPIRouter(logger, config, provider, ttime.Now(), repo)
	if err != nil {
		logger.Error("failed creating API router",
			"error", err,
		)
		os.Exit(1)
	}

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

	logger.Info("                                                     ")
	logger.Info("    /$$$$$$$$              /$$                       ")
	logger.Info("   |__  $$__/             | $$                       ")
	logger.Info("      | $$  /$$$$$$   /$$$$$$$  /$$$$$$   /$$$$$$$   ")
	logger.Info("      | $$ /$$__  $$ /$$__  $$ /$$__  $$ /$$_____/   ")
	logger.Info("      | $$| $$  ⧹ $$| $$  | $$| $$  ⧹ $$|  $$$$$$    ")
	logger.Info("      | $$| $$  | $$| $$  | $$| $$  | $$ ⧹____  $$   ")
	logger.Info("      | $$|  $$$$$$/|  $$$$$$$|  $$$$$$/ /$$$$$$$/   ")
	logger.Info("      |__/ ⧹______/  ⧹_______/ ⧹______/ |_______/    ")
	logger.Info("                                                     ")
	logger.Info("starting server",
		"service", config.Service.Name,
		"version", Version,
		"hostname", hostname,
		"location", config.Service.Location,
	)
	err = server.ListenAndServe()
	if err != nil {
		logger.Error("failed running server",
			"error", err,
		)
		os.Exit(1)
	}
}
