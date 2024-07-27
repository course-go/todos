[![Go Report Card](https://goreportcard.com/badge/github.com/course-go/todos)](https://goreportcard.com/report/github.com/course-go/todos)
![Go version](https://img.shields.io/github/go-mod/go-version/course-go/todos)
![CI status](https://github.com/course-go/todos/actions/workflows/ci-cd.yaml/badge.svg?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/course-go/todos/badge.svg)](https://coveralls.io/github/course-go/todos)

# Todos

Sample Todos web application.

## Packages

This project uses:

- [net/http](https://pkg.go.dev/net/http) for routing
- [pgx](https://github.com/jackc/pgx) for database access
- [migrate](https://github.com/jackc/pgx) for managing database migrations
- [slog](https://pkg.go.dev/log/slog) for logging
- [testcontainers](https://github.com/testcontainers/testcontainers-go) for testing with dependencies
