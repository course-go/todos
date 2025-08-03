COMPOSE_PROFILE?=all
VERSION=$(shell git describe --tags --dirty --abbrev=4 2> /dev/null || echo "0.0.0-devel")
LDFLAGS="-X main.Version=$(VERSION)"

.PHONY: all
all: build lint test

.PHONY: build
build:
	go build -ldflags $(LDFLAGS) -o bin/todos cmd/todos/main.go

.PHONY: run
run:
	go run cmd/todos/main.go

.PHONY: test
test:
	go test -cover -race -count=1 -timeout 300s -coverprofile=coverage.out ./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint run

.PHONY: clean
clean:
	rm -rf bin data coverage.out golangci-lint.out

.PHONY: dev
dev:
	docker compose --profile $(COMPOSE_PROFILE) up --build
