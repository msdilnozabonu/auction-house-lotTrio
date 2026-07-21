# Makefile шаблона. Главное — make help и make check.
#
# ── Windows ─────────────────────────────────────────────────
# Если `make` не работает: поставь Git Bash (идёт в Git for Windows),
# либо запускай команды вручную:
#   go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
#   go test ./...   &&   golangci-lint run ./...

GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)
NILAWAY       := $(shell command -v nilaway 2> /dev/null)

# установка инструментов, если их нет
setup:
	@if [ -z "$(GOLANGCI_LINT)" ]; then \
		echo "→ ставлю golangci-lint v2"; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest; \
	fi
	@if [ -z "$(NILAWAY)" ]; then \
		echo "→ ставлю nilaway"; \
		go install go.uber.org/nilaway/cmd/nilaway@latest; \
	fi

help: ## показать справку по командам
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

# На Windows к имени бинарника добавляем .exe.
BINEXT :=
ifeq ($(OS),Windows_NT)
	BINEXT := .exe
endif

build: ## собрать бинарник в bin/server
	go build -o bin/server$(BINEXT) ./cmd/server

run: ## запустить сервер (нужен поднятый Postgres — make db-up)
	go run ./cmd/server

db-up: ## поднять Postgres в Docker
	docker compose up -d

db-down: ## остановить Postgres
	docker compose down

test: ## unit-тесты
	go test ./...

cover: ## тесты + coverage (без cmd)
	go test -coverprofile=coverage.out $$(go list ./... | grep -v '/cmd/')
	@go tool cover -func=coverage.out | tail -n 1

race: ## тесты с race detector
	go test -race -timeout 60s ./...

fmt: ## форматирование
	go fmt ./...

vet: ## go vet ./...
	go vet ./...

tidy: ## go mod tidy
	go mod tidy

lint: setup ## golangci-lint + nilaway
	golangci-lint run ./...
	@if command -v nilaway >/dev/null 2>&1; then \
		nilaway -test=false ./... || true; \
	fi

fieldalignment: ## проверить/починить выравнивание полей структур
	@if ! command -v fieldalignment >/dev/null 2>&1; then \
		go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest; \
	fi
	fieldalignment -fix ./... || true

check: setup tidy fmt vet test lint ## ВСЁ: tidy + fmt + vet + tests + lint
	@echo "✓ All checks passed!"

clean: ## удалить временные файлы
	rm -rf bin/
	rm -f coverage.out coverage.xml

.PHONY: setup help build run db-up db-down test cover race fmt vet tidy lint fieldalignment check clean
.DEFAULT_GOAL := help
