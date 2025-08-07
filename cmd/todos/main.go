package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
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

func main() {
	err := runApp()
	if err != nil {
		slog.Error("failed running app",
			"err", err,
		)
		os.Exit(1)
	}
}

func runApp() error { //nolint: cyclop
	flag.Parse()

	if *versionFlag {
		fmt.Printf("TODOS: [%s]\n", Version) //nolint: forbidigo
		return nil
	}

	config, err := config.Parse(*configPathFlag)
	if err != nil {
		return fmt.Errorf("failed parsing config: %w", err)
	}

	location, err := time.LoadLocation(config.Location)
	if err != nil {
		return fmt.Errorf("failed loading location %s: %w", config.Location, err)
	}

	time.Local = location //nolint: gosmopolitan

	logger, err := logger.New(&config.Logging)
	if err != nil {
		return fmt.Errorf("failed creating logger: %w", err)
	}

	err = repository.Migrate(&config.Database, logger)
	if err != nil {
		return fmt.Errorf("failed migrating database: %w", err)
	}

	ctx := context.Background()

	registry, err := health.NewRegistry(ctx, health.WithService(config.Service.Name, Version))
	if err != nil {
		return fmt.Errorf("failed creating health registry: %w", err)
	}

	repo, err := repository.New(ctx, logger, registry, &config.Database)
	if err != nil {
		return fmt.Errorf("failed creating todo repository: %w", err)
	}

	exporter, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("failed creating prometheus exporter: %w", err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))

	metrics, err := metrics.New(provider)
	if err != nil {
		return fmt.Errorf("failed creating http metrics: %w", err)
	}

	hostname := net.JoinHostPort(config.Service.Host, config.Service.Port)
	validator := validator.New(validator.WithRequiredStructEnabled())
	todos := ctodos.NewController(validator, repo, ttime.Now())
	health := chealth.NewController(registry)

	server, err := http.NewServer(logger, metrics, hostname, health, todos)
	if err != nil {
		return fmt.Errorf("failed creating http server: %w", err)
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
		return fmt.Errorf("failed running http server: %w", err)
	}

	return nil
}
