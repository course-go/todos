[![Go Report Card](https://goreportcard.com/badge/github.com/course-go/todos)](https://goreportcard.com/report/github.com/course-go/todos)
![Go version](https://img.shields.io/github/go-mod/go-version/course-go/todos)
![CI status](https://github.com/course-go/todos/actions/workflows/ci-cd.yaml/badge.svg?branch=master)
[![SonarCloud Quality Gate](https://sonarcloud.io/api/project_badges/measure?project=course-go_todos&metric=alert_status)](https://sonarcloud.io/dashboard?id=course-go_todos)
[![Coverage Status](https://coveralls.io/repos/github/course-go/todos/badge.svg)](https://coveralls.io/github/course-go/todos)

# Todos

Sample Todos web application.

## Packages

This project uses:

- [net/http](https://pkg.go.dev/net/http) for routing
- [go-playground/validator](https://github.com/go-playground/validator) for input validation
- [pgx](https://github.com/jackc/pgx) for database access
- [migrate](https://github.com/jackc/pgx) for managing database migrations
- [slog](https://pkg.go.dev/log/slog) for logging
- [uuid](https://github.com/google/uuid) for IDs
- [go-cmp](https://github.com/google/go-cmp) for struct comparisons
- [testcontainers](https://github.com/testcontainers/testcontainers-go) for testing with dependencies
- [promethues/client_golang](github.com/prometheus/client_golang) for exporting Prometheus metrics
- [opentelemetry-go](go.opentelemetry.io/otel) for instrumenting telemetry
