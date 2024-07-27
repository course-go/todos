COMPOSE_PROFILE?=all
VERSION=$(shell git describe --tags --dirty --abbrev=4 || echo "0.0.0-devel")
LDFLAGS="-X main.Version=$(VERSION)"

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
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run

.PHONY: clean
clean:
	rm -rf bin data coverage.out

.PHONY: dev
dev:
	docker compose --profile $(COMPOSE_PROFILE) up --build
