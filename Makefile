-include .env
export

APP_PORT       ?= 8080
DB_HOST        ?= 127.0.0.1
DB_USER        ?= postgres
DB_NAME        ?= supplier_bridge
REDIS_ADDR     ?= localhost:6379

BINARY         := bin/api
DB_CONTAINER   := gopher-cart-db

.PHONY: run build mock \
        up down restart logs ps \
        psql migrate \
        lint test tidy \
        help

run:
	go run ./cmd/api

build:
	@mkdir -p bin
	go build -o $(BINARY) ./cmd/api

mock:
	go run ./cmd/mockserver

up:
	docker-compose up -d

down:
	docker-compose down -v

migrate:
	@for f in $$(ls migrations/*.sql | sort); do \
		echo "→ $$f"; \
		docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) < $$f; \
	done
