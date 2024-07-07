COMPOSE_PROFILE?=all

.PHONY: build
build:
	go build -o bin/todos cmd/todos/main.go

.PHONY: run
run:
	go run cmd/todos/main.go

.PHONY: test
test:
	go test -cover -race -count=1 -timeout 300s -coverprofile=coverage.out ./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run

.PHONY: dev
dev:
	docker compose --profile $(COMPOSE_PROFILE) up --build
