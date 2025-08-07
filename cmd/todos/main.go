package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/course-go/todos/internal/config"
	"github.com/course-go/todos/internal/health"
	"github.com/course-go/todos/internal/http"
	chealth "github.com/course-go/todos/internal/http/controllers/health"
	ctodos "github.com/course-go/todos/internal/http/controllers/todos"
	"github.com/course-go/todos/internal/http/metrics"
	"github.com/course-go/todos/internal/logger"
	"github.com/course-go/todos/internal/repository"
	ttime "github.com/course-go/todos/internal/time"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

var Version string

var (
	versionFlag    = flag.Bool("version", false, "output program version and exit")
	configPathFlag = flag.String("config", "/etc/course-go/todos/config.yaml", "path to config file")
)

func main() { //nolint: cyclop
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

	location, err := time.LoadLocation(config.Location)
	if err != nil {
		slog.Error("failed loading location",
			"error", err,
			"location", config.Location,
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

	registry, err := health.NewRegistry(ctx, health.WithService(config.Service.Name, Version))
	if err != nil {
		logger.Error("failed creating health registry",
			"error", err,
		)
		os.Exit(1)
	}

	repo, err := repository.New(ctx, logger, registry, &config.Database)
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

	metrics, err := metrics.New(provider)
	if err != nil {
		logger.Error("failed creating http metrics",
			"error", err,
		)
		os.Exit(1)
	}

	hostname := fmt.Sprintf("%s:%s",
		config.Service.Host,
		config.Service.Port,
	)
	validator := validator.New(validator.WithRequiredStructEnabled())
	todos := ctodos.NewTodosController(validator, repo, ttime.Now())
	health := chealth.NewHealthController(registry)

	server, err := http.NewServer(logger, metrics, hostname, health, todos)
	if err != nil {
		logger.Error("failed creating API router",
			"error", err,
		)
		os.Exit(1)
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
		"location", config.Location,
	)

	err = server.ListenAndServe()
	if err != nil {
		logger.Error("failed running server",
			"error", err,
		)
		os.Exit(1)
	}
}
