.PHONY: all help deps sqlc test build run clean docker-up docker-down migrate seed

APP_NAME=go-order-book
CMD_PATH=./cmd/main.go
BUILD_DIR=bin

all: deps sqlc test build

help:
	@echo "Comandos disponíveis:"
	@echo "  make deps        - Atualiza dependências com go mod tidy"
	@echo "  make sqlc        - Gera código do sqlc"
	@echo "  make test        - Roda todos os testes"
	@echo "  make build       - Compila o binário em $(BUILD_DIR)/$(APP_NAME)"
	@echo "  make run         - Roda a aplicação localmente"
	@echo "  make docker-up   - Sobe o Postgres com docker compose"
	@echo "  make docker-down - Para os containers"
	@echo "  make migrate     - Aplica migrations com tern"
	@echo "  make seed        - Insere dados iniciais para testes manuais"
	@echo "  make clean       - Remove artefatos de build"

deps:
	go mod tidy

sqlc:
	sqlc generate -f sqlc.yml

test:
	go test ./...

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "Build completo: $(BUILD_DIR)/$(APP_NAME)"

run:
	go run $(CMD_PATH)

docker-up:
	docker compose up -d db

docker-down:
	docker compose down

migrate:
	go run ./cmd/terndotenv

seed:
	docker compose exec -T db sh -c 'psql -U "$$DATABASE_USER" -d "$$DATABASE_NAME"' < scripts/seed.sql

clean:
	rm -rf $(BUILD_DIR)
