COMPOSE_PROFILE?=all

.PHONY: build
build:
	go build -o bin/todos cmd/todos/main.go


.PHONY: run
run:
	go run cmd/todos/main.go

.PHONY: dev
dev:
	docker compose --profile $(COMPOSE_PROFILE) up --build
